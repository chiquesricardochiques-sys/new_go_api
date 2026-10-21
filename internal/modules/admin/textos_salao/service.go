package textos_salao

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidData = errors.New(
	"dados inválidos",
)

type Service struct {
	repository *Repository
}

func NewService(
	repository *Repository,
) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Listar(
	ctx context.Context,
	req *ListRequest,
) ([]TextoSalao, error) {
	if req.SalaoID <= 0 {
		return nil, ErrInvalidData
	}

	return s.repository.Listar(
		ctx,
		req.SalaoID,
	)
}

func (s *Service) Buscar(
	ctx context.Context,
	req *GetRequest,
) (TextoSalao, error) {
	if req.SalaoID <= 0 {
		return TextoSalao{}, ErrInvalidData
	}

	chave := strings.TrimSpace(
		req.Chave,
	)

	if chave == "" ||
		len(chave) > 100 {
		return TextoSalao{}, ErrInvalidData
	}

	req.Chave = chave

	return s.repository.Buscar(
		ctx,
		req.SalaoID,
		req.Chave,
	)
}

func (s *Service) Adicionar(
	ctx context.Context,
	req *AddRequest,
) (TextoSalao, error) {
	if req.SalaoID <= 0 {
		return TextoSalao{}, ErrInvalidData
	}

	chave := strings.TrimSpace(
		strings.ToLower(req.Chave),
	)

	if chave == "" ||
		len(chave) > 100 {
		return TextoSalao{}, ErrInvalidData
	}

	conteudo := strings.TrimSpace(
		req.Conteudo,
	)

	if conteudo == "" {
		return TextoSalao{}, ErrInvalidData
	}

	req.Chave = chave
	req.Conteudo = conteudo

	return s.repository.Adicionar(
		ctx,
		req,
	)
}

func (s *Service) Atualizar(
	ctx context.Context,
	req *UpdateRequest,
) (TextoSalao, error) {
	if req.SalaoID <= 0 {
		return TextoSalao{}, ErrInvalidData
	}

	chave := strings.TrimSpace(
		strings.ToLower(req.Chave),
	)

	if chave == "" ||
		len(chave) > 100 {
		return TextoSalao{}, ErrInvalidData
	}

	conteudo := strings.TrimSpace(
		req.Conteudo,
	)

	if conteudo == "" {
		return TextoSalao{}, ErrInvalidData
	}

	req.Chave = chave
	req.Conteudo = conteudo

	return s.repository.Atualizar(
		ctx,
		req,
	)
}

func (s *Service) Excluir(
	ctx context.Context,
	req *DeleteRequest,
) error {
	if req.SalaoID <= 0 {
		return ErrInvalidData
	}

	chave := strings.TrimSpace(
		strings.ToLower(req.Chave),
	)

	if chave == "" ||
		len(chave) > 100 {
		return ErrInvalidData
	}

	req.Chave = chave

	return s.repository.Excluir(
		ctx,
		req,
	)
}