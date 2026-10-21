package salao

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidData = errors.New("dados inválidos")

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
	req *GetRequest,
) (Salao, error) {
	if req.SalaoID <= 0 {
		return Salao{}, ErrInvalidData
	}

	return s.repository.Buscar(
		ctx,
		req.SalaoID,
	)
}

func (s *Service) Atualizar(
	ctx context.Context,
	req *UpdateRequest,
) error {
	if req.SalaoID <= 0 {
		return ErrInvalidData
	}

	if req.Nome != nil {
		nome := strings.TrimSpace(*req.Nome)

		if nome == "" {
			return ErrInvalidData
		}

		if len(nome) > 150 {
			return ErrInvalidData
		}

		req.Nome = &nome
	}

	if req.Slug != nil {
		slug := strings.TrimSpace(
			strings.ToLower(*req.Slug),
		)

		if slug == "" {
			return ErrInvalidData
		}

		if len(slug) > 180 {
			return ErrInvalidData
		}

		req.Slug = &slug
	}

	if req.Email != nil {
		email := strings.TrimSpace(
			*req.Email,
		)

		if email != "" && len(email) > 255 {
			return ErrInvalidData
		}

		req.Email = &email
	}

	if req.Telefone != nil {
		telefone := strings.TrimSpace(
			*req.Telefone,
		)

		if len(telefone) > 30 {
			return ErrInvalidData
		}

		req.Telefone = &telefone
	}

	if req.Whatsapp != nil {
		whatsapp := strings.TrimSpace(
			*req.Whatsapp,
		)

		if len(whatsapp) > 30 {
			return ErrInvalidData
		}

		req.Whatsapp = &whatsapp
	}

	if req.Endereco != nil {
		endereco := strings.TrimSpace(
			*req.Endereco,
		)

		if len(endereco) > 255 {
			return ErrInvalidData
		}

		req.Endereco = &endereco
	}

	if req.Cidade != nil {
		cidade := strings.TrimSpace(
			*req.Cidade,
		)

		if len(cidade) > 100 {
			return ErrInvalidData
		}

		req.Cidade = &cidade
	}

	if req.Estado != nil {
		estado := strings.ToUpper(
			strings.TrimSpace(*req.Estado),
		)

		if estado != "" && len(estado) != 2 {
			return ErrInvalidData
		}

		req.Estado = &estado
	}

	if req.CEP != nil {
		cep := strings.TrimSpace(
			*req.CEP,
		)

		if len(cep) > 12 {
			return ErrInvalidData
		}

		req.CEP = &cep
	}

	if req.LogoURL != nil {
		logoURL := strings.TrimSpace(
			*req.LogoURL,
		)

		if len(logoURL) > 500 {
			return ErrInvalidData
		}

		req.LogoURL = &logoURL
	}

	if req.Status != nil &&
		!statusValido(*req.Status) {
		return ErrInvalidData
	}

	if req.Latitude != nil {
		if *req.Latitude < -90 ||
			*req.Latitude > 90 {
			return ErrInvalidData
		}
	}

	if req.Longitude != nil {
		if *req.Longitude < -180 ||
			*req.Longitude > 180 {
			return ErrInvalidData
		}
	}

	return s.repository.Atualizar(
		ctx,
		req,
	)
}

func statusValido(value string) bool {
	return value == "ativo" ||
		value == "inativo" ||
		value == "suspenso"
}