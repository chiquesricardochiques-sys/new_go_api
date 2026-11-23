package myschedule

import (
	"context"
	"errors"
	"time"
	_ "time/tzdata"
)

var (
	ErrDadosInvalidos          = errors.New("dados inválidos")
	ErrCenarioInvalido         = errors.New("cenário inválido")
	ErrStatusInvalido          = errors.New("status inválido")
	ErrAgendamentoNaoEncontrado = errors.New("agendamento não encontrado")
	ErrJaCancelado             = errors.New("agendamento já cancelado")
	ErrNaoPodeCancelar         = errors.New("agendamento não pode ser cancelado")
)

// ============================================================
// HORÁRIO DE BRASÍLIA
// ============================================================

var localBR = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("BRT", -3*60*60)
	}

	return loc
}()

var statusValidos = map[string]bool{
	"pendente":       true,
	"confirmado":    true,
	"cancelado":     true,
	"concluido":     true,
	"nao_compareceu": true,
}

// ============================================================
// SERVICE
// ============================================================

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// ============================================================
// LISTAR AGENDAMENTOS GLOBAIS
// ============================================================

func (s *Service) Listar(
	ctx context.Context,
	req ListarRequest,
) (*ListarData, error) {

	// Cliente é identificado globalmente.
	if req.AuthUserID <= 0 {
		return nil, ErrDadosInvalidos
	}

	cenario := req.Cenario

	if cenario == "" {
		cenario = CenarioMeusAgendamentos
	}

	if cenario != CenarioMeusAgendamentos &&
		cenario != CenarioHistorico {
		return nil, ErrCenarioInvalido
	}

	if req.Status != "" && !statusValidos[req.Status] {
		return nil, ErrStatusInvalido
	}

	agora := time.Now().In(localBR)

	inicioHoje := time.Date(
		agora.Year(),
		agora.Month(),
		agora.Day(),
		0, 0, 0, 0,
		localBR,
	)

	agendamentos, err := s.repository.ListarAgendamentos(
		ctx,
		req.AuthUserID,
		cenario,
		req.Status,
		inicioHoje,
	)

	if err != nil {
		return nil, err
	}

	for i := range agendamentos {
		a := &agendamentos[i]

		a.PodeCancelar =
			(a.Status == "pendente" || a.Status == "confirmado") &&
				a.Inicio.After(agora)
	}

	var status *string

	if req.Status != "" {
		st := req.Status
		status = &st
	}

	return &ListarData{
		Textos:       map[string]string{},
		Cenario:      cenario,
		Status:       status,
		Agendamentos: agendamentos,
	}, nil
}

// ============================================================
// CANCELAR AGENDAMENTO GLOBAL
// ============================================================

func (s *Service) Cancelar(
	ctx context.Context,
	req CancelarRequest,
) error {

	if req.AuthUserID <= 0 || req.AgendamentoID <= 0 {
		return ErrDadosInvalidos
	}

	return s.repository.CancelarAgendamento(
		ctx,
		req.AuthUserID,
		req.AgendamentoID,
		time.Now().In(localBR),
	)
}