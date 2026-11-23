package codeEnt

import (
	"context"
	"strings"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) BuscarSalao(
	ctx context.Context,
	codigo string,
) (*SalaoResponse, error) {

	codigo = strings.TrimSpace(codigo)

	if codigo == "" {
		return nil, ErrCodigoObrigatorio
	}

	salao, err := s.repository.BuscarPorCodigo(ctx, codigo)
	if err != nil {
		return nil, err
	}

	if salao.Status != "ativo" {
		return nil, ErrSalaoIndisponivel
	}

	return salao, nil
}