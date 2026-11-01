package painel

import (
	"context"
	"math"
	"strings"
	"time"
	_ "time/tzdata"
)

// "Hoje" sempre em horário de Brasília.
var localBR = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("BRT", -3*60*60)
	}
	return loc
}()

var rotulosStatus = map[string]string{
	"pendente":       "Pendente",
	"confirmado":     "Confirmado",
	"cancelado":      "Cancelado",
	"concluido":      "Concluído",
	"nao_compareceu": "Não compareceu",
}

var diasSemana = map[time.Weekday]string{
	time.Sunday:    "domingo",
	time.Monday:    "segunda-feira",
	time.Tuesday:   "terça-feira",
	time.Wednesday: "quarta-feira",
	time.Thursday:  "quinta-feira",
	time.Friday:    "sexta-feira",
	time.Saturday:  "sábado",
}

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func arredondar(valor float64) float64 {
	return math.Round(valor*100) / 100
}

// normalizarCodigo aceita "#0815332", "0815-332", " 0815332 " etc.
// e devolve só letras e números, em maiúsculas.
func normalizarCodigo(bruto string) (string, error) {
	codigo := strings.ToUpper(strings.TrimSpace(bruto))

	codigo = strings.NewReplacer(
		"#", "",
		" ", "",
		"-", "",
	).Replace(codigo)

	if len(codigo) < 4 || len(codigo) > 20 {
		return "", ErrDadosInvalidos
	}

	for _, c := range codigo {
		ehNumero := c >= '0' && c <= '9'
		ehLetra := c >= 'A' && c <= 'Z'

		if !ehNumero && !ehLetra {
			return "", ErrDadosInvalidos
		}
	}

	return codigo, nil
}

// resolverData: vazio = hoje; senão precisa ser YYYY-MM-DD.
func resolverData(texto string) (time.Time, error) {
	if strings.TrimSpace(texto) == "" {
		agora := time.Now().In(localBR)

		return time.Date(
			agora.Year(), agora.Month(), agora.Day(),
			0, 0, 0, 0, localBR,
		), nil
	}

	data, err := time.ParseInLocation("2006-01-02", texto, localBR)
	if err != nil {
		return time.Time{}, ErrDataInvalida
	}

	return data, nil
}

func (s *Service) Agenda(
	ctx context.Context,
	req AgendaRequest,
) (*AgendaData, error) {
	codigo, err := normalizarCodigo(req.Codigo)
	if err != nil {
		return nil, err
	}

	if req.ProfissionalID < 0 {
		return nil, ErrDadosInvalidos
	}

	data, err := resolverData(req.Data)
	if err != nil {
		return nil, err
	}

	salao, err := s.repository.BuscarSalaoPorCodigo(ctx, codigo)
	if err != nil {
		return nil, err
	}

	profissionais, err := s.repository.BuscarProfissionais(ctx, salao.ID)
	if err != nil {
		return nil, err
	}

	dataTexto := data.Format("2006-01-02")

	agendamentos, err := s.repository.ListarAgendamentos(
		ctx,
		salao.ID,
		dataTexto,
		req.ProfissionalID,
	)
	if err != nil {
		return nil, err
	}

	var resumo ResumoDia

	for i := range agendamentos {
		a := &agendamentos[i]

		a.PrecoTotal = arredondar(a.PrecoTotal)
		a.StatusLabel = rotulosStatus[a.Status]

		if a.StatusLabel == "" {
			a.StatusLabel = a.Status
		}

		resumo.Total++

		switch a.Status {
		case "pendente":
			resumo.Pendentes++
			resumo.ValorPrevisto += a.PrecoTotal
		case "confirmado":
			resumo.Confirmados++
			resumo.ValorPrevisto += a.PrecoTotal
		case "concluido":
			resumo.Concluidos++
			resumo.ValorPrevisto += a.PrecoTotal
		case "cancelado":
			resumo.Cancelados++
		case "nao_compareceu":
			resumo.NaoCompareceu++
		}
	}

	resumo.ValorPrevisto = arredondar(resumo.ValorPrevisto)

	hoje := time.Now().In(localBR).Format("2006-01-02")

	return &AgendaData{
		Salao:          *salao,
		Profissionais:  profissionais,
		Data:           dataTexto,
		DiaSemana:      diasSemana[data.Weekday()],
		Hoje:           dataTexto == hoje,
		ProfissionalID: req.ProfissionalID,
		Resumo:         resumo,
		Agendamentos:   agendamentos,
	}, nil
}