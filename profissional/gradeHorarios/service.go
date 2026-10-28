package gradehorarios

import (
	"context"
	"fmt"
	"strings"
	"time"
	_ "time/tzdata"
)

const (
	layoutData   = "2006-01-02"
	layoutSQL    = "2006-01-02 15:04:05"
	slotMinutos  = 30
	minutosNoDia = 24 * 60
)

var fusoBrasil = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("BRT", -3*60*60)
	}
	return loc
}()

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// ============================================================
// HELPERS
// ============================================================

func agora() time.Time {
	return time.Now().In(fusoBrasil)
}

func parseData(valor string) (time.Time, error) {
	t, err := time.ParseInLocation(layoutData, strings.TrimSpace(valor), fusoBrasil)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: data deve estar no formato YYYY-MM-DD", ErrDadosInvalidos)
	}
	return t, nil
}

func diaSemanaSigla(t time.Time) string {
	return [...]string{"dom", "seg", "ter", "qua", "qui", "sex", "sab"}[int(t.Weekday())]
}

// "HH:MM" -> minutos desde 00:00
func parseHora(valor string) (int, error) {
	t, err := time.Parse("15:04", strings.TrimSpace(valor))
	if err != nil {
		return 0, fmt.Errorf("%w: hora deve estar no formato HH:MM", ErrDadosInvalidos)
	}
	return t.Hour()*60 + t.Minute(), nil
}

func minutos(hhmm string) int {
	var h, m int
	fmt.Sscanf(hhmm, "%d:%d", &h, &m)
	return h*60 + m
}

func horaString(total int) string {
	return fmt.Sprintf("%02d:%02d", total/60, total%60)
}

func instante(dia time.Time, totalMinutos int) time.Time {
	return time.Date(dia.Year(), dia.Month(), dia.Day(), totalMinutos/60, totalMinutos%60, 0, 0, fusoBrasil)
}

// "YYYY-MM-DD HH:MM:SS" -> minutos desde 00:00 do dia `base`
// (parse em UTC dos dois lados: conta de relógio de parede, sem fuso/horário de verão).
func minutosDesde(base time.Time, dataHora string) int {
	t, err := time.Parse(layoutSQL, dataHora)
	if err != nil {
		return 0
	}
	b := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, time.UTC)
	return int(t.Sub(b).Minutes())
}

func (s *Service) profissionalDoSalao(ctx context.Context, b BaseRequest) (*Profissional, error) {
	if b.IDSalao <= 0 {
		return nil, fmt.Errorf("%w: id_salao obrigatório", ErrDadosInvalidos)
	}
	if b.ProfissionalID <= 0 {
		return nil, fmt.Errorf("%w: profissional_id obrigatório", ErrDadosInvalidos)
	}
	return s.repository.BuscarProfissional(ctx, b.IDSalao, b.ProfissionalID)
}

// ============================================================
// GRADE DO DIA (livre / agendado / bloqueado)
// ============================================================

type ocupacao struct {
	ini, fim int
}

// gerarSlots monta a grade de 30 em 30 minutos e marca cada slot.
// Um slot é "tocado" por um agendamento/bloqueio quando os intervalos se
// sobrepõem (agendamento de 10:15 às 10:45 ocupa os slots 10:00 e 10:30).
// minutoAgora >= 0: esconde slots que já começaram (só vale pra hoje).
func gerarSlots(
	g *Grade,
	dia time.Time,
	agendamentos []AgendamentoDia,
	bloqueios []BloqueioDia,
	minutoAgora int,
) []Slot {
	slots := make([]Slot, 0)

	abertura := minutos(g.Abertura)
	fechamento := minutos(g.Fechamento)

	pausaIni, pausaFim := -1, -1
	if g.PausaInicio != "" && g.PausaFim != "" {
		pausaIni, pausaFim = minutos(g.PausaInicio), minutos(g.PausaFim)
	}

	for m := abertura; m < fechamento; m += slotMinutos {
		if pausaIni >= 0 && m >= pausaIni && m < pausaFim {
			continue
		}
		if minutoAgora >= 0 && m <= minutoAgora {
			continue
		}

		slot := Slot{Hora: horaString(m), Status: StatusLivre, Disponivel: true}

		// 1) cliente agendado
		for _, ag := range agendamentos {
			ini := minutosDesde(dia, ag.Inicio)
			fim := ini + ag.Duracao

			if m < fim && m+slotMinutos > ini {
				id := ag.ID
				slot.Status = StatusAgendado
				slot.Disponivel = false
				slot.Agendamento = &id
				slot.Servicos = ag.Servicos
				break
			}
		}

		// 2) bloqueio do profissional
		if slot.Status == StatusLivre {
			for _, b := range bloqueios {
				ini := minutosDesde(dia, b.Inicio)
				fim := minutosDesde(dia, b.Fim)

				if m < fim && m+slotMinutos > ini {
					id := b.ID
					slot.Status = StatusBloqueado
					slot.Disponivel = false
					slot.BloqueioID = &id
					slot.Motivo = b.Motivo

					// mostra o bloqueio inteiro, limitado ao dia
					bi, bf := horaString(clamp(ini)), horaString(clamp(fim))
					slot.BloqueioInicio, slot.BloqueioFim = &bi, &bf
					break
				}
			}
		}

		slots = append(slots, slot)
	}

	return slots
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > minutosNoDia {
		return minutosNoDia
	}
	return v
}

func (s *Service) HorariosDia(ctx context.Context, req HorariosDiaRequest) (HorariosDiaResponse, error) {
	prof, err := s.profissionalDoSalao(ctx, req.BaseRequest)
	if err != nil {
		return HorariosDiaResponse{}, err
	}

	if strings.TrimSpace(req.Data) == "" {
		return HorariosDiaResponse{}, fmt.Errorf("%w: data obrigatória", ErrDadosInvalidos)
	}
	dia, err := parseData(req.Data)
	if err != nil {
		return HorariosDiaResponse{}, err
	}

	resposta := HorariosDiaResponse{UsaHorarios: prof.UsaHorarios, Horarios: []Slot{}}

	if prof.UsaHorarios == 0 {
		return resposta, nil
	}

	diaBloqueado, err := s.repository.DiaBloqueado(ctx, prof.ID, dia.Format(layoutData))
	if err != nil {
		return HorariosDiaResponse{}, err
	}
	if diaBloqueado {
		resposta.Bloqueado = true
		return resposta, nil
	}

	grade, err := s.repository.BuscarGrade(ctx, req.IDSalao, prof.ID, diaSemanaSigla(dia))
	if err != nil {
		return HorariosDiaResponse{}, err
	}
	if grade == nil || !grade.Abre {
		return resposta, nil
	}

	inicioDia := dia.Format(layoutSQL)
	fimDia := dia.AddDate(0, 0, 1).Format(layoutSQL)

	agendamentos, err := s.repository.AgendamentosNoIntervalo(ctx, req.IDSalao, prof.ID, inicioDia, fimDia)
	if err != nil {
		return HorariosDiaResponse{}, err
	}

	bloqueios, err := s.repository.BloqueiosNoIntervalo(ctx, prof.ID, inicioDia, fimDia)
	if err != nil {
		return HorariosDiaResponse{}, err
	}

	minutoAgora := -1
	agoraBR := agora()
	if dia.Format(layoutData) == agoraBR.Format(layoutData) {
		minutoAgora = agoraBR.Hour()*60 + agoraBR.Minute()
	}

	resposta.Horarios = gerarSlots(grade, dia, agendamentos, bloqueios, minutoAgora)
	return resposta, nil
}

// ============================================================
// BLOQUEAR
// ============================================================

// validarJanela confere se [ini, fim) cabe na grade: dentro do expediente,
// fora da pausa e alinhada aos slots de 30 minutos da grade.
func validarJanela(g *Grade, ini, fim int) error {
	if fim <= ini {
		return fmt.Errorf("%w: hora_fim deve ser maior que hora_inicio", ErrDadosInvalidos)
	}

	abertura := minutos(g.Abertura)
	fechamento := minutos(g.Fechamento)

	if (ini-abertura)%slotMinutos != 0 || (fim-ini)%slotMinutos != 0 {
		return fmt.Errorf("%w: os horários seguem a grade de %d minutos", ErrDadosInvalidos, slotMinutos)
	}

	pausaIni, pausaFim := -1, -1
	if g.PausaInicio != "" && g.PausaFim != "" {
		pausaIni, pausaFim = minutos(g.PausaInicio), minutos(g.PausaFim)
	}

	for m := ini; m < fim; m += slotMinutos {
		if m < abertura || m >= fechamento {
			return ErrForaDaGrade
		}
		if pausaIni >= 0 && m >= pausaIni && m < pausaFim {
			return ErrForaDaGrade
		}
	}

	return nil
}

func (s *Service) Bloquear(ctx context.Context, req BloquearRequest) (BloquearResponse, error) {
	prof, err := s.profissionalDoSalao(ctx, req.BaseRequest)
	if err != nil {
		return BloquearResponse{}, err
	}

	if prof.UsaHorarios == 0 {
		return BloquearResponse{}, ErrProfissionalSemHorarios
	}

	dia, err := parseData(req.Data)
	if err != nil {
		return BloquearResponse{}, err
	}

	ini, err := parseHora(req.HoraInicio)
	if err != nil {
		return BloquearResponse{}, err
	}
	fim, err := parseHora(req.HoraFim)
	if err != nil {
		return BloquearResponse{}, err
	}

	// Mesma regra da tela: horário que já começou não aparece (nem pode ser bloqueado).
	if fim > ini && !instante(dia, ini).After(agora()) {
		return BloquearResponse{}, ErrHorarioPassado
	}

	diaBloqueado, err := s.repository.DiaBloqueado(ctx, prof.ID, dia.Format(layoutData))
	if err != nil {
		return BloquearResponse{}, err
	}
	if diaBloqueado {
		return BloquearResponse{}, ErrDiaBloqueado
	}

	grade, err := s.repository.BuscarGrade(ctx, req.IDSalao, prof.ID, diaSemanaSigla(dia))
	if err != nil {
		return BloquearResponse{}, err
	}
	if grade == nil || !grade.Abre {
		return BloquearResponse{}, ErrForaDaGrade
	}

	if err := validarJanela(grade, ini, fim); err != nil {
		return BloquearResponse{}, err
	}

	// agendamento / bloqueio existente: conferido dentro da transação
	id, err := s.repository.CriarBloqueio(
		ctx,
		req.IDSalao,
		prof.ID,
		instante(dia, ini).Format(layoutSQL),
		instante(dia, fim).Format(layoutSQL),
		strings.TrimSpace(req.Motivo),
	)
	if err != nil {
		return BloquearResponse{}, err
	}

	return BloquearResponse{
		BloqueioID: id,
		Data:       dia.Format(layoutData),
		HoraInicio: horaString(ini),
		HoraFim:    horaString(fim),
	}, nil
}

// ============================================================
// DESBLOQUEAR
// ============================================================

func (s *Service) Desbloquear(ctx context.Context, req DesbloquearRequest) (DesbloquearResponse, error) {
	if _, err := s.profissionalDoSalao(ctx, req.BaseRequest); err != nil {
		return DesbloquearResponse{}, err
	}

	if req.BloqueioID <= 0 {
		return DesbloquearResponse{}, fmt.Errorf("%w: bloqueio_id obrigatório", ErrDadosInvalidos)
	}

	if err := s.repository.RemoverBloqueio(ctx, req.IDSalao, req.ProfissionalID, req.BloqueioID); err != nil {
		return DesbloquearResponse{}, err
	}

	return DesbloquearResponse{BloqueioID: req.BloqueioID}, nil
}