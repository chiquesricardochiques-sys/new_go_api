
package professionals

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
	return &Service{repository: repository}
}

// List retorna todos os profissionais ou um profissional específico.
func (s *Service) List(
	ctx context.Context,
	req *ListRequest,
) ([]Professional, error) {
	if req.SalaoID <= 0 || req.ID < 0 {
		return nil, ErrInvalidData
	}

	// ID igual a zero significa que a consulta é uma listagem geral.
	if req.ID == 0 {
		return s.repository.Listar(ctx, req.SalaoID)
	}

	profissional, err := s.repository.Buscar(
		ctx,
		req.SalaoID,
		req.ID,
	)
	if err != nil {
		return nil, err
	}

	return []Professional{profissional}, nil
}

// Update atualiza um profissional já existente.
func (s *Service) Update(
	ctx context.Context,
	req *UpdateRequest,
) error {
	if req.SalaoID <= 0 || req.ID <= 0 {
		return ErrInvalidData
	}

	if req.Nome != nil {
		nome := strings.TrimSpace(*req.Nome)
		if nome == "" {
			return ErrInvalidData
		}
		req.Nome = &nome
	}

	if req.Status != nil && !statusValido(*req.Status) {
		return ErrInvalidData
	}

	if req.TipoRemuneracao != nil &&
		!remuneracaoValida(*req.TipoRemuneracao) {
		return ErrInvalidData
	}

	if req.FrequenciaPagamento != nil &&
		!frequenciaValida(*req.FrequenciaPagamento) {
		return ErrInvalidData
	}

	if req.PercentualComissao != nil &&
		(*req.PercentualComissao < 0 ||
			*req.PercentualComissao > 100) {
		return ErrInvalidData
	}

	if req.SalarioFixo != nil && *req.SalarioFixo < 0 {
		return ErrInvalidData
	}

	if req.Especialidade != nil {
		especialidade := strings.TrimSpace(*req.Especialidade)
		req.Especialidade = &especialidade
	}

	return s.repository.Atualizar(ctx, req)
}

func statusValido(value string) bool {
	return value == "ativo" ||
		value == "inativo" ||
		value == "aguardando_liberacao"
}

func remuneracaoValida(value string) bool {
	return value == "comissao" ||
		value == "fixo" ||
		value == "misto"
}

func frequenciaValida(value string) bool {
	return value == "semanal" ||
		value == "quinzenal" ||
		value == "mensal"
}
