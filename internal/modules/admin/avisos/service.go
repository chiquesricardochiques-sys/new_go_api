package avisos

import (
	"context"
	"errors"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Listar(
	ctx context.Context,
	salaoID int64,
	historico bool,
) (*ListarAvisosResponse, error) {

	if salaoID <= 0 {
		return nil, errors.New("ID do salão inválido")
	}

	avisos, err := s.repository.Listar(
		ctx,
		salaoID,
		historico,
	)
	if err != nil {
		return nil, err
	}

	return &ListarAvisosResponse{
		Avisos: avisos,
		Total:  len(avisos),
	}, nil
}