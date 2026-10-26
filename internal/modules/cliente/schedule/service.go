package schedule

import (
	"context"
	"errors"
	"fmt"
	"time"
	"sort"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func validarData(data string) (time.Time, error) {
	if data == "" {
		return time.Time{}, ErrDataInvalida
	}

	// A data é interpretada no fuso local do servidor, o mesmo usado
	// para montar hora_marcada e para o "hoje" da tela.
	dataParsed, err := time.ParseInLocation(
		"2006-01-02",
		data,
		time.Local,
	)
	if err != nil {
		return time.Time{}, ErrDataInvalida
	}

	return dataParsed, nil
}

func validarHora(hora string) (time.Time, error) {
	if hora == "" {
		return time.Time{}, ErrHoraInvalida
	}

	horaParsed, err := time.Parse(
		"15:04",
		hora,
	)
	if err != nil {
		return time.Time{}, ErrHoraInvalida
	}

	return horaParsed, nil
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
	case time.Saturday:
		return "sab"
	default:
		return ""
	}
}

func horaParaMinutos(valor string) int {
	if len(valor) < 5 {
		return 0
	}

	hora := int(valor[0]-'0')*10 +
		int(valor[1]-'0')

	minuto := int(valor[3]-'0')*10 +
		int(valor[4]-'0')

	return hora*60 + minuto
}

func minutosParaHora(minutos int) string {
	hora := minutos / 60
	minuto := minutos % 60

	return fmt.Sprintf(
		"%02d:%02d",
		hora,
		minuto,
	)
}

func intervaloSobrepoe(
	inicio time.Time,
	fim time.Time,
	agendamento AgendamentoExistente,
) bool {
	return inicio.Before(agendamento.Fim) &&
		fim.After(agendamento.Inicio)
}

// func (s *Service) BuscarHorarios(
// 	ctx context.Context,
// 	req HorariosDisponiveisRequest,
// ) (int, []Horario, error) {
// 	if req.SalaoID <= 0 ||
// 		req.ProfissionalID <= 0 {
// 		return 0, nil, ErrDadosInvalidos
// 	}

// 	data, err := validarData(req.Data)
// 	if err != nil {
// 		return 0, nil, err
// 	}

// 	profissional, err := s.repository.BuscarProfissional(
// 		ctx,
// 		req.SalaoID,
// 		req.ProfissionalID,
// 	)
// 	if err != nil {
// 		return 0, nil, err
// 	}

// 	if !profissional.UsaHorarios {
// 		return 0, []Horario{}, nil
// 	}

// 	diaBloqueado, err :=
// 		s.repository.VerificarDiaBloqueado(
// 			ctx,
// 			req.ProfissionalID,
// 			req.Data,
// 		)
// 	if err != nil {
// 		return 0, nil, err
// 	}

// 	if diaBloqueado {
// 		return 1, []Horario{}, nil
// 	}

// 	grade, err := s.repository.BuscarGrade(
// 		ctx,
// 		req.SalaoID,
// 		req.ProfissionalID,
// 		diaSemana(data),
// 	)

// 	if errors.Is(err, ErrGradeNaoEncontrada) {
// 		return 1, []Horario{}, nil
// 	}

// 	if err != nil {
// 		return 0, nil, err
// 	}

// 	if !grade.Abre {
// 		return 1, []Horario{}, nil
// 	}

// 	agendamentos, err :=
// 		s.repository.BuscarAgendamentosDoDia(
// 			ctx,
// 			req.SalaoID,
// 			req.ProfissionalID,
// 			req.Data,
// 		)
// 	if err != nil {
// 		return 0, nil, err
// 	}

// 	abertura := horaParaMinutos(
// 		grade.Abertura,
// 	)

// 	fechamento := horaParaMinutos(
// 		grade.Fechamento,
// 	)

// 	pausaInicio := -1
// 	pausaFim := -1

// 	if grade.PausaInicio != nil &&
// 		grade.PausaFim != nil {
// 		pausaInicio = horaParaMinutos(
// 			*grade.PausaInicio,
// 		)

// 		pausaFim = horaParaMinutos(
// 			*grade.PausaFim,
// 		)
// 	}

// 	horarios := make([]Horario, 0)

// 	for minuto := abertura; minuto < fechamento; minuto += 30 {
// 		if pausaInicio >= 0 &&
// 			pausaFim >= 0 &&
// 			minuto >= pausaInicio &&
// 			minuto < pausaFim {
// 			continue
// 		}

// 		horarios = append(
// 			horarios,
// 			Horario{
// 				Hora:       minutosParaHora(minuto),
// 				Disponivel: true,
// 				Servicos:   nil,
// 			},
// 		)
// 	}

// 	// Remove horários que já passaram quando a data é hoje.
// 	agora := time.Now()

// 	if data.Format("2006-01-02") ==
// 		agora.Format("2006-01-02") {

// 		minutoAtual :=
// 			agora.Hour()*60 +
// 				agora.Minute()

// 		filtrados := make(
// 			[]Horario,
// 			0,
// 			len(horarios),
// 		)

// 		for _, horario := range horarios {
// 			minutoHorario :=
// 				horaParaMinutos(horario.Hora)

// 			if minutoHorario > minutoAtual {
// 				filtrados = append(
// 					filtrados,
// 					horario,
// 				)
// 			}
// 		}

// 		horarios = filtrados
// 	}

// 	// Marca horários ocupados pelos agendamentos existentes
// 	// (cada agendamento ocupa hora_marcada até hora_marcada + duracao_minutos).
// 	for i := range horarios {
// 		horarioMinutos :=
// 			horaParaMinutos(
// 				horarios[i].Hora,
// 			)

// 		inicio := time.Date(
// 			data.Year(),
// 			data.Month(),
// 			data.Day(),
// 			horarioMinutos/60,
// 			horarioMinutos%60,
// 			0,
// 			0,
// 			time.Local,
// 		)

// 		// Cada slot representa 30 minutos.
// 		fim := inicio.Add(30 * time.Minute)

// 		for _, agendamento :=
// 			range agendamentos {

// 			if intervaloSobrepoe(
// 				inicio,
// 				fim,
// 				agendamento,
// 			) {
// 				horarios[i].Disponivel = false

// 				id := agendamento.ID
// 				horarios[i].AgendamentoID = &id

// 				break
// 			}
// 		}
// 	}

// 	return 1, horarios, nil
// }
func (s *Service) BuscarHorarios(
	ctx context.Context,
	req HorariosDisponiveisRequest,
) (int, []Horario, error) {

	if req.SalaoID <= 0 ||
		req.ProfissionalID <= 0 {
		return 0, nil, ErrDadosInvalidos
	}

	data, err := validarData(req.Data)
	if err != nil {
		return 0, nil, err
	}

	profissional, err := s.repository.BuscarProfissional(
		ctx,
		req.SalaoID,
		req.ProfissionalID,
	)
	if err != nil {
		return 0, nil, err
	}

	if !profissional.UsaHorarios {
		return 0, []Horario{}, nil
	}

	diaBloqueado, err :=
		s.repository.VerificarDiaBloqueado(
			ctx,
			req.ProfissionalID,
			req.Data,
		)
	if err != nil {
		return 0, nil, err
	}

	if diaBloqueado {
		return 1, []Horario{}, nil
	}

	grade, err := s.repository.BuscarGrade(
		ctx,
		req.SalaoID,
		req.ProfissionalID,
		diaSemana(data),
	)

	if errors.Is(err, ErrGradeNaoEncontrada) {
		return 1, []Horario{}, nil
	}

	if err != nil {
		return 0, nil, err
	}

	if !grade.Abre {
		return 1, []Horario{}, nil
	}

	agendamentos, err :=
		s.repository.BuscarAgendamentosDoDia(
			ctx,
			req.SalaoID,
			req.ProfissionalID,
			req.Data,
		)
	if err != nil {
		return 0, nil, err
	}

	abertura := horaParaMinutos(
		grade.Abertura,
	)

	fechamento := horaParaMinutos(
		grade.Fechamento,
	)

	pausaInicio := -1
	pausaFim := -1

	if grade.PausaInicio != nil &&
		grade.PausaFim != nil {

		pausaInicio = horaParaMinutos(
			*grade.PausaInicio,
		)

		pausaFim = horaParaMinutos(
			*grade.PausaFim,
		)
	}

	// ============================================================
	// MONTA OS PONTOS DE HORÁRIO
	//
	// Os 30 minutos continuam sendo a grade principal.
	// Porém, o final de um agendamento também vira um ponto
	// de horário disponível.
	//
	// Exemplo:
	//
	// 08:00 -> 08:40
	//
	// Grade:
	// 08:00
	// 08:30
	// 09:00
	//
	// Autoajuste:
	// 08:40
	//
	// Resultado:
	// 08:00
	// 08:30
	// 08:40
	// 09:00
	// ============================================================

	pontos := make(map[int]struct{})

	// Grade normal de 30 em 30 minutos.
	for minuto := abertura; minuto < fechamento; minuto += 30 {

		if pausaInicio >= 0 &&
			pausaFim >= 0 &&
			minuto >= pausaInicio &&
			minuto < pausaFim {
			continue
		}

		pontos[minuto] = struct{}{}
	}

	// Adiciona o final real dos agendamentos como novo ponto.
	for _, agendamento := range agendamentos {

		minutoFim :=
			agendamento.Fim.Hour()*60 +
				agendamento.Fim.Minute()

		// O final precisa estar dentro do expediente.
		if minutoFim < abertura ||
			minutoFim >= fechamento {
			continue
		}

		// Não cria horário dentro da pausa.
		if pausaInicio >= 0 &&
			pausaFim >= 0 &&
			minutoFim >= pausaInicio &&
			minutoFim < pausaFim {
			continue
		}

		pontos[minutoFim] = struct{}{}
	}

	// Converte os pontos para slice.
	minutos := make(
		[]int,
		0,
		len(pontos),
	)

	for minuto := range pontos {
		minutos = append(
			minutos,
			minuto,
		)
	}

	// Mantém os horários em ordem cronológica.
	sort.Ints(minutos)

	horarios := make(
		[]Horario,
		0,
		len(minutos),
	)

	for _, minuto := range minutos {

		horarios = append(
			horarios,
			Horario{
				Hora:       minutosParaHora(minuto),
				Disponivel: true,
				Servicos:   nil,
			},
		)
	}

	// ============================================================
	// REMOVE HORÁRIOS QUE JÁ PASSARAM QUANDO A DATA É HOJE
	// ============================================================

	agora := time.Now()

	if data.Format("2006-01-02") ==
		agora.Format("2006-01-02") {

		minutoAtual :=
			agora.Hour()*60 +
				agora.Minute()

		filtrados := make(
			[]Horario,
			0,
			len(horarios),
		)

		for _, horario := range horarios {

			minutoHorario :=
				horaParaMinutos(
					horario.Hora,
				)

			if minutoHorario > minutoAtual {
				filtrados = append(
					filtrados,
					horario,
				)
			}
		}

		horarios = filtrados
	}

	// ============================================================
	// MARCA OS PONTOS QUE ESTÃO DENTRO DE UM AGENDAMENTO
	//
	// Aqui NÃO usamos mais:
	//
	//     inicio + 30 minutos
	//
	// Porque 30 minutos é apenas a grade visual.
	//
	// O agendamento real usa:
	//
	//     agendamento.Inicio -> agendamento.Fim
	//
	// Exemplo:
	//
	// 08:00 -> 08:40
	//
	// 08:00 = indisponível
	// 08:30 = indisponível
	// 08:40 = disponível
	// ============================================================

	for i := range horarios {

		horarioMinutos :=
			horaParaMinutos(
				horarios[i].Hora,
			)

		inicio := time.Date(
			data.Year(),
			data.Month(),
			data.Day(),
			horarioMinutos/60,
			horarioMinutos%60,
			0,
			0,
			time.Local,
		)

		for _, agendamento :=
			range agendamentos {

			// O ponto está ocupado se estiver dentro
			// do intervalo real do agendamento.
			if !inicio.Before(agendamento.Inicio) &&
				inicio.Before(agendamento.Fim) {

				horarios[i].Disponivel = false

				id := agendamento.ID
				horarios[i].AgendamentoID = &id

				break
			}
		}
	}

	return 1, horarios, nil
}

func (s *Service) SchedulePage(
	ctx context.Context,
	req SchedulePageRequest,
) (*ScheduleData, error) {
	if req.SalaoID <= 0 {
		return nil, ErrDadosInvalidos
	}

	if err := s.repository.VerificarSalao(
		ctx,
		req.SalaoID,
	); err != nil {
		return nil, err
	}

	agora := time.Now()
	hoje := agora.Format("2006-01-02")

	// Data escolhida: a que veio na requisição (se válida e não passada) ou hoje.
	dataSelecionada := hoje

	if req.Data != "" {
		data, err := validarData(req.Data)
		if err != nil {
			return nil, err
		}

		inicioDeHoje := time.Date(
			agora.Year(),
			agora.Month(),
			agora.Day(),
			0, 0, 0, 0,
			time.Local,
		)

		if data.Before(inicioDeHoje) {
			return nil, ErrDataInvalida
		}

		dataSelecionada = req.Data
	}

	textos, err := s.repository.BuscarTextos(
		ctx,
		req.SalaoID,
	)
	if err != nil {
		return nil, err
	}

	profissionais, err :=
		s.repository.BuscarProfissionais(
			ctx,
			req.SalaoID,
		)
	if err != nil {
		return nil, err
	}

	// Profissional escolhido: o da requisição ou o primeiro da lista.
	var selecionado *Profissional

	if req.ProfissionalID > 0 {
		for i := range profissionais {
			if profissionais[i].ID == req.ProfissionalID {
				selecionado = &profissionais[i]
				break
			}
		}

		if selecionado == nil {
			return nil, ErrProfissionalNaoEncontrado
		}
	} else if len(profissionais) > 0 {
		selecionado = &profissionais[0]
	}

	servicos := make([]Servico, 0)
	horarios := make([]Horario, 0)

	var profissionalSelecionado *int64
	var usaHorarios int

	if selecionado != nil {
		id := selecionado.ID
		profissionalSelecionado = &id

		// Só os serviços que este profissional faz.
		servicos, err =
			s.repository.BuscarServicosDoProfissional(
				ctx,
				req.SalaoID,
				id,
			)
		if err != nil {
			return nil, err
		}

		usaHorarios, horarios, err =
			s.BuscarHorarios(
				ctx,
				HorariosDisponiveisRequest{
					SalaoID:        req.SalaoID,
					ProfissionalID: id,
					Data:           dataSelecionada,
				},
			)
		if err != nil {
			return nil, err
		}
	}

	return &ScheduleData{
		Textos:               textos,
		Servicos:             servicos,
		Profissionais:        profissionais,
		CurrentDate:          hoje,
		UsaHorarios:          usaHorarios,
		HorariosDisponiveis:  horarios,
		SelectedProfissional: profissionalSelecionado,
		SelectedData:         dataSelecionada,
	}, nil
}

// ServicosDoProfissional devolve só os serviços que o profissional faz.
// Usado quando o cliente troca de profissional na tela de agendamento.
func (s *Service) ServicosDoProfissional(
	ctx context.Context,
	req ServicosProfissionalRequest,
) ([]Servico, error) {
	if req.SalaoID <= 0 ||
		req.ProfissionalID <= 0 {
		return nil, ErrDadosInvalidos
	}

	if _, err := s.repository.BuscarProfissional(
		ctx,
		req.SalaoID,
		req.ProfissionalID,
	); err != nil {
		return nil, err
	}

	return s.repository.BuscarServicosDoProfissional(
		ctx,
		req.SalaoID,
		req.ProfissionalID,
	)
}

func (s *Service) Agendar(
	ctx context.Context,
	req AgendarRequest,
) (int64, error) {
	if req.SalaoID <= 0 ||
		req.IDCliente <= 0 ||
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

	hora, err := validarHora(req.Hora)
	if err != nil {
		return 0, err
	}

	agora := time.Now()

	inicioDeHoje := time.Date(
		agora.Year(),
		agora.Month(),
		agora.Day(),
		0, 0, 0, 0,
		time.Local,
	)

	if data.Before(inicioDeHoje) {
		return 0, ErrDataInvalida
	}

	profissional, err :=
		s.repository.BuscarProfissional(
			ctx,
			req.SalaoID,
			req.ProfissionalID,
		)
	if err != nil {
		return 0, err
	}

	if !profissional.UsaHorarios {
		return 0, ErrProfissionalSemHorarios
	}

	// Confere o cliente (uma única consulta).
	authUserID, err :=
		s.repository.BuscarClienteAuthUserID(
			ctx,
			req.IDCliente,
		)
	if err != nil {
		return 0, err
	}

	// Confere que os serviços são do profissional e soma a duração.
	servicos, duracaoTotal, err :=
		s.repository.BuscarServicosSelecionados(
			ctx,
			req.SalaoID,
			req.ProfissionalID,
			req.ServicoIDs,
		)
	if err != nil {
		return 0, err
	}

	if duracaoTotal <= 0 {
		return 0, ErrServicoNaoEncontrado
	}

	grade, err :=
		s.repository.BuscarGrade(
			ctx,
			req.SalaoID,
			req.ProfissionalID,
			diaSemana(data),
		)
	if errors.Is(err, ErrGradeNaoEncontrada) {
		return 0, ErrHorarioIndisponivel
	}

	if err != nil {
		return 0, err
	}

	if !grade.Abre {
		return 0, ErrHorarioIndisponivel
	}

	minutoInicio :=
		hora.Hour()*60 +
			hora.Minute()

	minutoAbertura :=
		horaParaMinutos(grade.Abertura)

	minutoFechamento :=
		horaParaMinutos(grade.Fechamento)

	if minutoInicio < minutoAbertura {
		return 0, ErrHorarioIndisponivel
	}

	minutoFim :=
		minutoInicio + duracaoTotal

	if minutoFim > minutoFechamento {
		return 0, ErrHorarioIndisponivel
	}

	if grade.PausaInicio != nil &&
		grade.PausaFim != nil {

		pausaInicio :=
			horaParaMinutos(*grade.PausaInicio)

		pausaFim :=
			horaParaMinutos(*grade.PausaFim)

		if minutoInicio < pausaFim &&
			minutoFim > pausaInicio {
			return 0, ErrHorarioIndisponivel
		}
	}

	// Data e hora escolhidas pelo cliente (vai para hora_marcada).
	horaMarcada := time.Date(
		data.Year(),
		data.Month(),
		data.Day(),
		hora.Hour(),
		hora.Minute(),
		0,
		0,
		time.Local,
	)

	// Fim previsto: só para conferir bloqueios de horário.
	fimPrevisto := horaMarcada.Add(
		time.Duration(duracaoTotal) * time.Minute,
	)

	// Hoje só vale para horários que ainda não passaram.
	if data.Format("2006-01-02") ==
		agora.Format("2006-01-02") {

		if !horaMarcada.After(agora) {
			return 0, ErrHorarioIndisponivel
		}
	}

	diaBloqueado, err :=
		s.repository.VerificarDiaBloqueado(
			ctx,
			req.ProfissionalID,
			req.Data,
		)
	if err != nil {
		return 0, err
	}

	if diaBloqueado {
		return 0, ErrDiaBloqueado
	}

	horarioBloqueado, err :=
		s.repository.VerificarHorarioBloqueado(
			ctx,
			req.ProfissionalID,
			horaMarcada,
			fimPrevisto,
		)
	if err != nil {
		return 0, err
	}

	if horarioBloqueado {
		return 0, ErrHorarioBloqueado
	}

	return s.repository.CriarAgendamento(
		ctx,
		req,
		authUserID,
		servicos,
		horaMarcada,
		duracaoTotal,
	)
}