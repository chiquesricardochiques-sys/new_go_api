package dadosagendamentos

import (
	"context"
	"fmt"
	"math"
	"time"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
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

func arredondar(valor float64) float64 {
	return math.Round(valor*100) / 100
}

// nomeDiaSemana recebe "YYYY-MM-DD" e devolve "segunda-feira", etc.
func nomeDiaSemana(data string) string {
	parsed, err := time.Parse("2006-01-02", data)
	if err != nil {
		return ""
	}

	return diasSemana[parsed.Weekday()]
}

// calcularHoraFim soma a duração à hora marcada ("HH:MM").

func calcularHoraFim(hora string, duracaoMinutos int) string {
	if hora == "" || duracaoMinutos <= 0 {
		return ""
	}

	inicio, err := time.Parse("15:04", hora)
	if err != nil {
		return ""
	}

	fim := inicio.Add(
		time.Duration(duracaoMinutos) * time.Minute,
	)

	return fim.Format("15:04")
}


// Buscar devolve os dados completos de UM agendamento pendente do cliente.
func (s *Service) Buscar(
	ctx context.Context,
	req DadosRequest,
) (*AgendamentoDados, error) {

	if req.SalaoID <= 0 ||
		req.AgendamentoID <= 0 ||
		req.IDCliente <= 0 {
		return nil, fmt.Errorf(
			"%w: salao_id, agendamento_id e id_cliente são obrigatórios",
			ErrDadosInvalidos,
		)
	}

	dados, err := s.repository.BuscarAgendamento(
		ctx,
		&req,
	)
	if err != nil {
		return nil, err
	}

	/*
	 * Esta tela continua sendo a tela
	 * de detalhes do agendamento pendente.
	 */
	if dados.Status != "pendente" {
		return nil, ErrNaoPendente
	}

	servicos, err := s.repository.BuscarServicos(
		ctx,
		dados.ID,
	)
	if err != nil {
		return nil, err
	}

	total := 0.0

	for i := range servicos {

		servicos[i].Subtotal = arredondar(
			servicos[i].PrecoUnitario *
				float64(servicos[i].Quantidade),
		)

		total += servicos[i].Subtotal
	}

	dados.Servicos = servicos

	dados.PrecoTotal =
		arredondar(total)

	dados.StatusLabel =
		"Pendente"

	dados.DiaSemana =
		nomeDiaSemana(dados.Data)

	dados.HoraFim =
		calcularHoraFim(
			dados.Hora,
			dados.DuracaoMinutos,
		)

	return dados, nil
}