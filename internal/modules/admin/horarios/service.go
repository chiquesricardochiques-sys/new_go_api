package horarios

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrInvalidData = errors.New("dados inválidos")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Listar(
	ctx context.Context,
	req ListRequest,
) ([]Horario, error) {
	if req.IDSalao <= 0 || req.ProfissionalID <= 0 {
		return nil, ErrInvalidData
	}

	if _, err := time.Parse("2006-01-02", req.Data); err != nil {
		return nil, ErrInvalidData
	}

	return s.repo.Listar(
		ctx,
		req.IDSalao,
		req.ProfissionalID,
		req.Data,
	)
}

func (s *Service) Criar(
	ctx context.Context,
	req CreateRequest,
) error {
	req.Data = strings.TrimSpace(req.Data)
	req.Hora = strings.TrimSpace(req.Hora)

	if req.IDSalao <= 0 || req.ProfissionalID <= 0 {
		return ErrInvalidData
	}

	if _, err := time.Parse("2006-01-02", req.Data); err != nil {
		return ErrInvalidData
	}

	if _, err := time.Parse("15:04", req.Hora); err != nil {
		if _, errSegundos := time.Parse("15:04:05", req.Hora); errSegundos != nil {
			return ErrInvalidData
		}
	}

	return s.repo.Criar(
		ctx,
		req.IDSalao,
		req.ProfissionalID,
		req.Data,
		req.Hora,
	)
}

func (s *Service) Excluir(
	ctx context.Context,
	req DeleteRequest,
) error {
	if req.IDSalao <= 0 || req.ID <= 0 {
		return ErrInvalidData
	}

	return s.repo.Excluir(ctx, req.IDSalao, req.ID)
}
