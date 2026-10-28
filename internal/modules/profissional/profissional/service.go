package profissional

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

func (s *Service) AddRelations(
	ctx context.Context,
	req AddRelationsRequest,
) error {
	if req.IDSalao <= 0 ||
		req.AuthUserID <= 0 ||
		strings.TrimSpace(req.Nome) == "" {
		return ErrDadosInvalidos
	}

	req.Nome = strings.TrimSpace(req.Nome)

	return s.repository.AddRelations(ctx, req)
}

func (s *Service) Verify(
	ctx context.Context,
	req VerifyRequest,
) (VerifyResponse, error) {
	if req.IDSalao <= 0 || req.AuthUserID <= 0 {
		return VerifyResponse{}, ErrDadosInvalidos
	}

	return s.repository.Verify(ctx, req)
}