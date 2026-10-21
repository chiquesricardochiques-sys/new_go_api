package schedule

import (
	"context"
	"errors"
	"fmt"
	"time"
	_ "time/tzdata" // garante America/Sao_Paulo mesmo em container sem tzdata
)

var (
	ErrSalaoNaoEncontrado        = errors.New("salão não encontrado")
	ErrProfissionalNaoEncontrado = errors.New("profissional não encontrado")
	ErrClienteNaoEncontrado      = errors.New("cliente não encontrado")
	ErrServicoNaoEncontrado      = errors.New("serviço não encontrado")
	ErrGradeNaoEncontrada        = errors.New("grade não encontrada")
	ErrDiaBloqueado              = errors.New("dia bloqueado")
	ErrHorarioBloqueado          = errors.New("horário bloqueado")
	ErrHorarioIndisponivel       = errors.New("horário indisponível")
	ErrDataInvalida              = errors.New("data inválida")
	ErrHoraInvalida              = errors.New("hora inválida")
	ErrDadosInvalidos            = errors.New("dados inválidos")
	ErrProfissionalSemHorarios   = errors.New("profissional não trabalha com horários")
)

// Cada slot da grade tem 30 minutos.
const intervaloSlot = 30

// Fuso do salão. O "hoje" e o "agora" são sempre em horário de Brasília
// (o Node usava UTC e virava o dia às 21h).
var localBR = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("BRT", -3*60*60)
	}
	return loc
}()

func agoraBR() time.Time {
	return time.Now().In(localBR)
}

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func validarData(data string) (time.Time, error) {
	if data == "" {
		return time.Time{}, ErrDataInvalida
	}
	t, err := time.ParseInLocation("2006-01-02", data, localBR)
	if err != nil {
		return time.Time{}, ErrDataInvalida
	}
	return t, nil
}

// Aceita "15:04" e "15:04:05". Retorna minutos desde 00:00.
func validarHora(hora string) (int, error) {
	if hora == "" {
		return 0, ErrHoraInvalida
	}
	t, err := time.Parse("15:04", hora)
	if err != nil {
		t, err = time.Parse("15:04:05", hora)
		if err != nil {
			return 0, ErrHoraInvalida
		}
	}
	return t.Hour()*60 + t.Minute(), nil
}

func diaSemana(data time.Time) string {
	switch data.Weekday() {
	case time.Sunday:
		return "dom"
	case time.Monday:
		return "seg"
	case time.Tuesday:
		return "ter"
	case time.Wednesday:
		return "qua"
	case time.Thursday:
		return "qui"
	case time.Friday:
		return "sex"
	default:
		return "sab"
	}
}

// "08:00" ou "08:00:00" -> 480
func horaParaMinutos(valor string) int {
	if len(valor) < 5 {
		return 0
	}
	return int(valor[0]-'0')*600 + int(valor[1]-'0')*60 +
		int(valor[3]-'0')*10 + int(valor[4]-'0')
}

func minutosParaHora(minutos int) string {
	return fmt.Sprintf("%02d:%02d", minutos/60, minutos%60)
}

func instante(dia time.Time, minutos int) time.Time {
	return time.Date(dia.Year(), dia.Month(), dia.Day(), minutos/60, minutos%60, 0, 0, localBR)
}

func sobrepoe(aIni, aFim, bIni, bFim time.Time) bool {
	return aIni.Before(bFim) && aFim.After(bIni)
}

// ============================================================
// HORÁRIOS DA GRADE (com os já agendados marcados)
// ============================================================

func (s *Service) BuscarHorarios(ctx context.Context, req HorariosDisponiveisRequest) (int, []Horario, error) {
	if req.SalaoID <= 0 || req.ProfissionalID <= 0 {
		return 0, nil, ErrDadosInvalidos
	}

	data, err := validarData(req.Data)
	if err != nil {
		return 0, nil, err
	}

	profissional, err := s.repository.BuscarProfissional(ctx, req.SalaoID, req.ProfissionalID)
	if err != nil {
		return 0, nil, err
	}

	if !profissional.UsaHorarios {
		return 0, []Horario{}, nil
	}

	diaBloqueado, err := s.repository.VerificarDiaBloqueado(ctx, req.ProfissionalID, req.Data)
	if err != nil {
		return 0, nil, err
	}
	if diaBloqueado {
		return 1, []Horario{}, nil
	}

	grade, err := s.repository.BuscarGrade(ctx, req.SalaoID, req.ProfissionalID, diaSemana(data))
	if errors.Is(err, ErrGradeNaoEncontrada) {
		return 1, []Horario{}, nil
	}
	if err != nil {
		return 0, nil, err
	}
	if !grade.Abre {
		return 1, []Horario{}, nil
	}

	inicioDia := data
	fimDia := data.AddDate(0, 0, 1)

	agendamentos, err := s.repository.BuscarAgendamentosDoDia(ctx, req.SalaoID, req.ProfissionalID, inicioDia, fimDia)
	if err != nil {
		return 0, nil, err
	}

	bloqueios, err := s.repository.BuscarBloqueiosDoDia(ctx, req.ProfissionalID, inicioDia, fimDia)
	if err != nil {
		return 0, nil, err
	}

	abertura := horaParaMinutos(grade.Abertura)
	fechamento := horaParaMinutos(grade.Fechamento)

	pausaInicio, pausaFim := -1, -1
	if grade.PausaInicio != nil && grade.PausaFim != nil {
		pausaInicio = horaParaMinutos(*grade.PausaInicio)
		pausaFim = horaParaMinutos(*grade.PausaFim)
	}

	// hoje: só mostra horários que ainda não passaram
	agora := agoraBR()
	ehHoje := data.Format("2006-01-02") == agora.Format("2006-01-02")
	minutoAgora := agora.Hour()*60 + agora.Minute()

	horarios := make([]Horario, 0)

	for minuto := abertura; minuto < fechamento; minuto += intervaloSlot {
		// pausa (almoço etc.)
		if pausaInicio >= 0 && minuto >= pausaInicio && minuto < pausaFim {
			continue
		}
		if ehHoje && minuto <= minutoAgora {
			continue
		}

		slotIni := instante(data, minuto)
		slotFim := slotIni.Add(intervaloSlot * time.Minute)

		h := Horario{Hora: minutosParaHora(minuto), Disponivel: true}

		// 1) já tem gente agendada nesse horário?
		for _, ag := range agendamentos {
			if sobrepoe(slotIni, slotFim, ag.Inicio, ag.Fim) {
				id := ag.ID
				h.Disponivel = false
				h.Motivo = "agendado"
				h.AgendamentoID = &id
				h.Servicos = ag.Servicos
				break
			}
		}

		// 2) bloqueio de horário do profissional
		if h.Disponivel {
			for _, b := range bloqueios {
				if sobrepoe(slotIni, slotFim, b.Inicio, b.Fim) {
					h.Disponivel = false
					h.Motivo = "bloqueado"
					break
				}
			}
		}

		horarios = append(horarios, h)
	}

	return 1, horarios, nil
}

// ============================================================
// TELA DE AGENDAMENTO
// ============================================================

func (s *Service) SchedulePage(ctx context.Context, req SchedulePageRequest) (*ScheduleData, error) {
	if req.SalaoID <= 0 {
		return nil, ErrDadosInvalidos
	}

	if err := s.repository.VerificarSalao(ctx, req.SalaoID); err != nil {
		return nil, err
	}

	// data escolhida (ou hoje)
	dataSelecionada := agoraBR().Format("2006-01-02")
	if req.Data != "" {
		if _, err := validarData(req.Data); err != nil {
			return nil, err
		}
		dataSelecionada = req.Data
	}

	textos, err := s.repository.BuscarTextos(ctx, req.SalaoID)
	if err != nil {
		return nil, err
	}

	servicos, err := s.repository.BuscarServicos(ctx, req.SalaoID)
	if err != nil {
		return nil, err
	}

	profissionais, err := s.repository.BuscarProfissionais(ctx, req.SalaoID)
	if err != nil {
		return nil, err
	}

	// profissional escolhido (ou o primeiro)
	var profissionalSelecionado *int64
	if req.ProfissionalID > 0 {
		for _, p := range profissionais {
			if p.ID == req.ProfissionalID {
				id := p.ID
				profissionalSelecionado = &id
				break
			}
		}
		if profissionalSelecionado == nil {
			return nil, ErrProfissionalNaoEncontrado
		}
	} else if len(profissionais) > 0 {
		id := profissionais[0].ID
		profissionalSelecionado = &id
	}

	usaHorarios := 1
	horarios := []Horario{}

	if profissionalSelecionado != nil {
		usaHorarios, horarios, err = s.BuscarHorarios(ctx, HorariosDisponiveisRequest{
			SalaoID:        req.SalaoID,
			ProfissionalID: *profissionalSelecionado,
			Data:           dataSelecionada,
		})
		if err != nil {
			return nil, err
		}
	}

	return &ScheduleData{
		Textos:               textos,
		Servicos:             servicos,
		Profissionais:        profissionais,
		CurrentDate:          agoraBR().Format("2006-01-02"),
		UsaHorarios:          usaHorarios,
		HorariosDisponiveis:  horarios,
		SelectedProfissional: profissionalSelecionado,
		SelectedData:         dataSelecionada,
	}, nil
}

// ============================================================
// AGENDAR
// ============================================================

func (s *Service) Agendar(ctx context.Context, req AgendarRequest) (int64, error) {
	if req.SalaoID <= 0 ||
		req.AuthUserID <= 0 ||
		req.ProfissionalID <= 0 ||
		req.Data == "" ||
		req.Hora == "" ||
		len(req.ServicoIDs) == 0 {
		return 0, ErrDadosInvalidos
	}

	data, err := validarData(req.Data)
	if err != nil {
		return 0, err
	}

	minutoInicio, err := validarHora(req.Hora)
	if err != nil {
		return 0, err
	}

	agora := agoraBR()
	hoje := time.Date(agora.Year(), agora.Month(), agora.Day(), 0, 0, 0, 0, localBR)
	if data.Before(hoje) {
		return 0, ErrDataInvalida
	}

	profissional, err := s.repository.BuscarProfissional(ctx, req.SalaoID, req.ProfissionalID)
	if err != nil {
		return 0, err
	}
	if !profissional.UsaHorarios {
		return 0, ErrProfissionalSemHorarios
	}

	if err := s.repository.VerificarCliente(ctx, req.AuthUserID); err != nil {
		return 0, err
	}

	servicos, duracaoTotal, err := s.repository.BuscarServicosSelecionados(
		ctx, req.SalaoID, req.ProfissionalID, req.ServicoIDs)
	if err != nil {
		return 0, err
	}
	if duracaoTotal <= 0 {
		return 0, ErrServicoNaoEncontrado
	}

	// --- o horário precisa caber na grade do profissional ---
	grade, err := s.repository.BuscarGrade(ctx, req.SalaoID, req.ProfissionalID, diaSemana(data))
	if errors.Is(err, ErrGradeNaoEncontrada) {
		return 0, ErrHorarioIndisponivel
	}
	if err != nil {
		return 0, err
	}
	if !grade.Abre {
		return 0, ErrHorarioIndisponivel
	}

	abertura := horaParaMinutos(grade.Abertura)
	fechamento := horaParaMinutos(grade.Fechamento)
	minutoFim := minutoInicio + duracaoTotal

	if minutoInicio < abertura || minutoFim > fechamento {
		return 0, ErrHorarioIndisponivel
	}

	// só aceita horários que existem na grade (de 30 em 30 a partir da abertura)
	if (minutoInicio-abertura)%intervaloSlot != 0 {
		return 0, ErrHoraInvalida
	}

	if grade.PausaInicio != nil && grade.PausaFim != nil {
		pausaInicio := horaParaMinutos(*grade.PausaInicio)
		pausaFim := horaParaMinutos(*grade.PausaFim)
		if minutoInicio < pausaFim && minutoFim > pausaInicio {
			return 0, ErrHorarioIndisponivel
		}
	}

	inicio := instante(data, minutoInicio)
	fim := instante(data, minutoFim)

	if !inicio.After(agora) {
		return 0, ErrHorarioIndisponivel
	}

	// --- bloqueios do profissional ---
	diaBloqueado, err := s.repository.VerificarDiaBloqueado(ctx, req.ProfissionalID, req.Data)
	if err != nil {
		return 0, err
	}
	if diaBloqueado {
		return 0, ErrDiaBloqueado
	}

	horarioBloqueado, err := s.repository.VerificarHorarioBloqueado(ctx, req.ProfissionalID, inicio, fim)
	if err != nil {
		return 0, err
	}
	if horarioBloqueado {
		return 0, ErrHorarioBloqueado
	}

	// --- já tem gente agendada? (checagem final, dentro da transação) ---
	return s.repository.CriarAgendamento(ctx, req, servicos, inicio, fim)
}