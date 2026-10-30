package home

import (
	"context"
	"errors"
)

var (
	ErrDadosInvalidos = errors.New(
		"salao_id é obrigatório",
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
	req HomeRequest,
) (*HomeData, error) {
	if req.SalaoID <= 0 {
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

	return &HomeData{
		Textos:        textos,
		Salao:         salao,
		Profissionais: profissionais,
	}, nil
}