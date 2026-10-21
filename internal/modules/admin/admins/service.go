
package admins

import (
	"context"
	"errors"
	"fmt"
)

var ErrDadosInvalidos = errors.New("dados inválidos")

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func validarIDs(idSalao, authUserID int64) error {
	if idSalao <= 0 {
		return fmt.Errorf("%w: id_salao obrigatório", ErrDadosInvalidos)
	}

	if authUserID <= 0 {
		return fmt.Errorf("%w: auth_user_id obrigatório", ErrDadosInvalidos)
	}

	return nil
}

func (s *Service) AddRelations(
	ctx context.Context,
	req AddRelationsRequest,
) error {
	if err := validarIDs(req.IDSalao, req.AuthUserID); err != nil {
		return err
	}

	return s.repository.AddRelations(
		ctx,
		req.IDSalao,
		req.AuthUserID,
	)
}

func (s *Service) Verify(
	ctx context.Context,
	req VerifyRequest,
) (VerifyResponse, error) {
	if err := validarIDs(req.IDSalao, req.AuthUserID); err != nil {
		return VerifyResponse{}, err
	}

	return s.repository.Verify(
		ctx,
		req.IDSalao,
		req.AuthUserID,
	)
}
