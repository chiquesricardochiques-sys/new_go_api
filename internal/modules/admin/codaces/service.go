package codaces

import (
	"errors"
	"strings"
)

var ErrSalaoNaoEncontrado = errors.New("salão não encontrado")

var ErrCodigoDuplicado = errors.New("código já existente")

var ErrCodigoNaoCadastrado = errors.New("código não cadastrado")

var ErrCodigoInvalido = errors.New("código inválido")

var ErrSalaoJaPossuiCodigo = errors.New("salão já possui código")

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// ------------------------------------------------------------
// BUSCAR CÓDIGO
// ------------------------------------------------------------

func (s *Service) BuscarCodigo(
	salaoID uint64,
) (*CodigoResponse, error) {

	if salaoID == 0 {
		return nil, ErrSalaoNaoEncontrado
	}

	return s.repository.BuscarCodigo(salaoID)
}

// ------------------------------------------------------------
// INSERIR NOVO CÓDIGO
// ------------------------------------------------------------

func (s *Service) InserirCodigo(
	salaoID uint64,
	codigo string,
) error {

	codigo = strings.TrimSpace(codigo)

	if salaoID == 0 {
		return ErrSalaoNaoEncontrado
	}

	if codigo == "" {
		return ErrCodigoInvalido
	}

	// Verifica se o salão existe.
	existe, err := s.repository.SalaoExiste(salaoID)
	if err != nil {
		return err
	}

	if !existe {
		return ErrSalaoNaoEncontrado
	}

	// Verifica se o salão já possui código.
	_, err = s.repository.BuscarCodigo(salaoID)

	if err == nil {
		return ErrSalaoJaPossuiCodigo
	}

	if !errors.Is(err, ErrCodigoNaoCadastrado) {
		if errors.Is(err, ErrSalaoNaoEncontrado) {
			return ErrSalaoNaoEncontrado
		}

		return err
	}

	// Verifica se o código pertence a outro salão.
	duplicado, err := s.repository.CodigoExisteEmOutroSalao(
		codigo,
		salaoID,
	)

	if err != nil {
		return err
	}

	if duplicado {
		return ErrCodigoDuplicado
	}

	// Insere o código.
	return s.repository.InserirCodigo(salaoID, codigo)
}

// ------------------------------------------------------------
// ALTERAR CÓDIGO
// ------------------------------------------------------------

func (s *Service) AlterarCodigo(
	salaoID uint64,
	codigo string,
) error {

	codigo = strings.TrimSpace(codigo)

	if salaoID == 0 {
		return ErrSalaoNaoEncontrado
	}

	if codigo == "" {
		return ErrCodigoInvalido
	}

	// Verifica se o salão existe.
	existe, err := s.repository.SalaoExiste(salaoID)
	if err != nil {
		return err
	}

	if !existe {
		return ErrSalaoNaoEncontrado
	}

	// Verifica se o salão possui código cadastrado.
	_, err = s.repository.BuscarCodigo(salaoID)

	if errors.Is(err, ErrCodigoNaoCadastrado) {
		return ErrCodigoNaoCadastrado
	}

	if errors.Is(err, ErrSalaoNaoEncontrado) {
		return ErrSalaoNaoEncontrado
	}

	if err != nil {
		return err
	}

	// Verifica se o código pertence a outro salão.
	duplicado, err := s.repository.CodigoExisteEmOutroSalao(
		codigo,
		salaoID,
	)

	if err != nil {
		return err
	}

	if duplicado {
		return ErrCodigoDuplicado
	}

	// Altera o código existente.
	return s.repository.AlterarCodigo(salaoID, codigo)
}