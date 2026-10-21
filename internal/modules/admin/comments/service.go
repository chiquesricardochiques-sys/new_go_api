package comments

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidData = errors.New("dados inválidos")

type ImageService interface {
	Upload(ctx context.Context, filePath string, idSalao int64) (*CommentImage, error)
	Update(ctx context.Context, filePath string, oldPublicID string, idSalao int64) (*CommentImage, error)
	Delete(ctx context.Context, publicID string) error
}

type Service struct {
	repo   *Repository
	images ImageService
}

func NewService(repo *Repository, images ImageService) *Service {
	return &Service{
		repo:   repo,
		images: images,
	}
}

func (s *Service) Listar(ctx context.Context, idSalao int64) ([]Comment, error) {
	if idSalao <= 0 {
		return nil, ErrInvalidData
	}

	return s.repo.Listar(ctx, idSalao)
}

func (s *Service) Criar(
	ctx context.Context,
	req CreateRequest,
	filePath string,
) error {
	req.Nome = strings.TrimSpace(req.Nome)
	req.Comentario = strings.TrimSpace(req.Comentario)

	if req.IDSalao <= 0 || req.Nome == "" || req.Comentario == "" {
		return ErrInvalidData
	}

	var imagem CommentImage

	if filePath != "" {
		if s.images == nil {
			return errors.New("serviço de imagens não configurado")
		}

		result, err := s.images.Upload(ctx, filePath, req.IDSalao)
		if err != nil {
			return err
		}
		if result != nil {
			imagem = *result
		}
	}

	return s.repo.Criar(
		ctx,
		req.IDSalao,
		req.Nome,
		req.Comentario,
		imagem,
	)
}

func (s *Service) Atualizar(
	ctx context.Context,
	req UpdateRequest,
	filePath string,
) error {
	req.Nome = strings.TrimSpace(req.Nome)
	req.Comentario = strings.TrimSpace(req.Comentario)

	if req.IDSalao <= 0 || req.ID <= 0 ||
		req.Nome == "" || req.Comentario == "" {
		return ErrInvalidData
	}

	_, publicID, err := s.repo.Buscar(ctx, req.IDSalao, req.ID)
	if err != nil {
		return err
	}

	var imagem *CommentImage

	if filePath != "" {
		if s.images == nil {
			return errors.New("serviço de imagens não configurado")
		}

		result, err := s.images.Update(
			ctx,
			filePath,
			publicID,
			req.IDSalao,
		)
		if err != nil {
			return err
		}
		imagem = result
	}

	return s.repo.Atualizar(
		ctx,
		req.IDSalao,
		req.ID,
		req.Nome,
		req.Comentario,
		req.Ativo,
		imagem,
	)
}

func (s *Service) Excluir(
	ctx context.Context,
	req DeleteRequest,
) error {
	if req.IDSalao <= 0 || req.ID <= 0 {
		return ErrInvalidData
	}

	_, publicID, err := s.repo.Buscar(ctx, req.IDSalao, req.ID)
	if err != nil {
		return err
	}

	if err := s.repo.Excluir(ctx, req.IDSalao, req.ID); err != nil {
		return err
	}

	if publicID != "" && s.images != nil {
		// A exclusão do comentário já ocorreu.
		// Falha ao remover a imagem não deve restaurar o comentário.
		_ = s.images.Delete(ctx, publicID)
	}

	return nil
}
