package cliente

import (
	"context"
	"errors"
)

var (
	ErrDadosInvalidos = errors.New(
		"auth_user_id e salao_id são obrigatórios",
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

func validarRequest(req RelacaoRequest) error {
	if req.AuthUserID <= 0 || req.SalaoID <= 0 {
		return ErrDadosInvalidos
	}

	return nil
}

// Verificar somente consulta a relação existente.
func (s *Service) Verificar(
	ctx context.Context,
	req RelacaoRequest,
) (*RelacaoResponse, error) {
	if err := validarRequest(req); err != nil {
		return nil, err
	}

	// O salão enviado pelo site/app precisa existir
	// e estar ativo.
	if err := s.repository.VerificarSalao(
		ctx,
		req.SalaoID,
	); err != nil {
		return nil, err
	}

	idCliente, err := s.repository.BuscarRelacao(
		ctx,
		req.AuthUserID,
	)

	if err != nil {
		return nil, err
	}

	return &RelacaoResponse{
		IDCliente: idCliente,
	}, nil
}

// Adicionar cria a relação caso não exista.
// Se já existir, retorna o mesmo id_cliente.
func (s *Service) Adicionar(
	ctx context.Context,
	req RelacaoRequest,
) (*RelacaoResponse, bool, error) {
	if err := validarRequest(req); err != nil {
		return nil, false, err
	}

	// Primeiro valida o salão.
	if err := s.repository.VerificarSalao(
		ctx,
		req.SalaoID,
	); err != nil {
		return nil, false, err
	}

	idCliente, criado, err := s.repository.AdicionarRelacao(
		ctx,
		req.AuthUserID,
	)

	if err != nil {
		return nil, false, err
	}
	return &RelacaoResponse{
		IDCliente: idCliente,
	}, criado, nil
}