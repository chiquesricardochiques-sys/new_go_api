package client_home

import (
	"context"
	"errors"
	"time"
	_ "time/tzdata"
)

var (
	ErrDadosInvalidos = errors.New(
		"salao_id e auth_user_id são obrigatórios",
	)
)

// "Agora" sempre em horário de Brasília.
var localBR = func() *time.Location {
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
	return &Service{
		repository: repository,
	}
}

func (s *Service) Buscar(
	ctx context.Context,
	req ClientHomeRequest,
) (*ClientHomeData, error) {
	if req.SalaoID <= 0 || req.AuthUserID <= 0 {
		return nil, ErrDadosInvalidos
	}

	salao, err := s.repository.BuscarSalao(
		ctx,
		req.SalaoID,
	)
	if err != nil {
		return nil, err
	}

	textos, err := s.repository.BuscarTextos(
		ctx,
		req.SalaoID,
	)
	if err != nil {
		return nil, err
	}

	profissionais, err := s.repository.BuscarProfissionais(
		ctx,
		req.SalaoID,
	)
	if err != nil {
		return nil, err
	}

	proximoAgendamento, err := s.repository.BuscarProximoAgendamento(
		ctx,
		req.SalaoID,
		req.AuthUserID,
		time.Now().In(localBR),
	)
	if err != nil {
		return nil, err
	}

	return &ClientHomeData{
		Textos:             textos,
		Salao:              salao,
		Profissionais:      profissionais,
		ProximoAgendamento: proximoAgendamento,
	}, nil
}