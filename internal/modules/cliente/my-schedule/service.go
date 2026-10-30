package myschedule

import (
	"context"
	"errors"
	"time"
	_ "time/tzdata"
)

var (
	ErrDadosInvalidos           = errors.New("dados inválidos")
	ErrCenarioInvalido          = errors.New("cenário inválido")
	ErrStatusInvalido           = errors.New("status inválido")
	ErrAgendamentoNaoEncontrado = errors.New("agendamento não encontrado")
	ErrJaCancelado              = errors.New("agendamento já cancelado")
	ErrNaoPodeCancelar          = errors.New("agendamento não pode ser cancelado")
)

// "Hoje" e "agora" sempre em horário de Brasília.
var localBR = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("BRT", -3*60*60)
	}
	return loc
}()

var statusValidos = map[string]bool{
	"pendente":       true,
	"confirmado":     true,
	"cancelado":      true,
	"concluido":      true,
	"nao_compareceu": true,
}

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Listar(ctx context.Context, req ListarRequest) (*ListarData, error) {
	if req.SalaoID <= 0 || req.AuthUserID <= 0 {
		return nil, ErrDadosInvalidos
	}

	cenario := req.Cenario
	if cenario == "" {
		cenario = CenarioMeusAgendamentos
	}
	if cenario != CenarioMeusAgendamentos && cenario != CenarioHistorico {
		return nil, ErrCenarioInvalido
	}

	if req.Status != "" && !statusValidos[req.Status] {
		return nil, ErrStatusInvalido
	}

	agora := time.Now().In(localBR)
	inicioHoje := time.Date(agora.Year(), agora.Month(), agora.Day(), 0, 0, 0, 0, localBR)

	textos, err := s.repository.BuscarTextos(ctx, req.SalaoID)
	if err != nil {
		return nil, err
	}

	agendamentos, err := s.repository.ListarAgendamentos(
		ctx, req.SalaoID, req.AuthUserID, cenario, req.Status, inicioHoje)
	if err != nil {
		return nil, err
	}

	// o botão "cancelar" só aparece pra quem ainda pode cancelar
	for i := range agendamentos {
		a := &agendamentos[i]
		a.PodeCancelar = (a.Status == "pendente" || a.Status == "confirmado") &&
			a.Inicio.After(agora)
	}

	var status *string
	if req.Status != "" {
		st := req.Status
		status = &st
	}

	return &ListarData{
		Textos:       textos,
		Cenario:      cenario,
		Status:       status,
		Agendamentos: agendamentos,
	}, nil
}

func (s *Service) Cancelar(ctx context.Context, req CancelarRequest) error {
	if req.SalaoID <= 0 || req.AuthUserID <= 0 || req.AgendamentoID <= 0 {
		return ErrDadosInvalidos
	}

	return s.repository.CancelarAgendamento(
		ctx, req.SalaoID, req.AuthUserID, req.AgendamentoID, time.Now().In(localBR))
}