package conf_prof

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	ChavePodeVerRendimento = "pode_ver_rendimento"
)

var (
	ErrProfissionalInvalido = errors.New(
		"profissional inválido",
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

// PodeVerRendimento consulta a configuração.
//
// REGRA:
// - se não existir configuração:
//   cria automaticamente como true
// - se existir:
//   retorna o valor salvo
func (s *Service) PodeVerRendimento(
	ctx context.Context,
	profissionalID int64,
) (bool, error) {

	if profissionalID <= 0 {
		return false, ErrProfissionalInvalido
	}

	valor, err := s.repository.BuscarValor(
		ctx,
		profissionalID,
		ChavePodeVerRendimento,
	)

	if err != nil {
		return false, fmt.Errorf(
			"buscar configuração pode ver rendimento: %w",
			err,
		)
	}

	// =====================================================
	// CONFIGURAÇÃO AINDA NÃO EXISTE
	// =====================================================
	//
	// Primeira vez:
	// cria automaticamente como SIM.
	//
	if valor == nil {
		permitido := true

		err := s.repository.DefinirValor(
			ctx,
			profissionalID,
			ChavePodeVerRendimento,
			permitido,
		)

		if err != nil {
			return false, fmt.Errorf(
				"criar configuração padrão pode ver rendimento: %w",
				err,
			)
		}

		return permitido, nil
	}

	// =====================================================
	// CONFIGURAÇÃO JÁ EXISTE
	// =====================================================

	var permitido bool

	if err := json.Unmarshal(
		valor,
		&permitido,
	); err != nil {
		return false, fmt.Errorf(
			"interpretar configuração %s: %w",
			ChavePodeVerRendimento,
			err,
		)
	}

	return permitido, nil
}

// DefinirPodeVerRendimento altera a configuração.
//
// Se não existir:
// INSERT.
//
// Se já existir:
// UPDATE.
func (s *Service) DefinirPodeVerRendimento(
	ctx context.Context,
	profissionalID int64,
	permitido bool,
) error {

	if profissionalID <= 0 {
		return ErrProfissionalInvalido
	}

	err := s.repository.DefinirValor(
		ctx,
		profissionalID,
		ChavePodeVerRendimento,
		permitido,
	)

	if err != nil {
		return fmt.Errorf(
			"salvar configuração pode ver rendimento: %w",
			err,
		)
	}

	return nil
}