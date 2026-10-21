package client_home

import (
	"context"
	"errors"
)

var (
	ErrDadosInvalidos = errors.New(
		"salao_id e id_cliente são obrigatórios",
	)
)

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
	if req.SalaoID <= 0 || req.IDCliente <= 0 {
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
		req.IDCliente,
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