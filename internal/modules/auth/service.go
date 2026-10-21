package auth

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDadosInvalidos = errors.New("email e senha são obrigatórios")
	ErrCredenciais    = errors.New("email ou senha incorretos")
	ErrProfInativo    = errors.New("este profissional está inativo")
	ErrSalaoInvalido  = errors.New("id_salao inválido")
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func validarLogin(req LoginRequest) error {
	if req.IDSalao <= 0 {
		return ErrSalaoInvalido
	}

	if strings.TrimSpace(req.Email) == "" ||
		req.Senha == "" {
		return ErrDadosInvalidos
	}

	return nil
}

func senhaCorreta(senha, hash string) bool {
	if hash == "" {
		return false
	}

	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(senha),
	) == nil
}

func (s *Service) LoginAdmin(
	ctx context.Context,
	req LoginRequest,
) (*AdminResponse, error) {
	if err := validarLogin(req); err != nil {
		return nil, err
	}

	admin, err := s.repository.BuscarAdmin(
		ctx, req.IDSalao, strings.TrimSpace(req.Email),
	)
	if err != nil {
		if errors.Is(err, ErrUsuarioNaoEncontrado) {
			return nil, ErrCredenciais
		}
		return nil, err
	}

	if !senhaCorreta(req.Senha, admin.SenhaHash) {
		return nil, ErrCredenciais
	}

	return &AdminResponse{
		ID:       admin.ID,
		Nome:     admin.Nome,
		Email:    admin.Email,
		CriadoEm: admin.CriadoEm,
	}, nil
}

func (s *Service) LoginProfissional(
	ctx context.Context,
	req LoginRequest,
) (*ProfissionalResponse, error) {
	if err := validarLogin(req); err != nil {
		return nil, err
	}

	profissional, err := s.repository.BuscarProfissional(
		ctx, req.IDSalao, strings.TrimSpace(req.Email),
	)
	if err != nil {
		if errors.Is(err, ErrUsuarioNaoEncontrado) {
			return nil, ErrCredenciais
		}
		return nil, err
	}

	if !profissional.Ativo {
		return nil, ErrProfInativo
	}

	if !senhaCorreta(req.Senha, profissional.SenhaHash) {
		return nil, ErrCredenciais
	}

	return &ProfissionalResponse{
		ID:            profissional.ID,
		Nome:          profissional.Nome,
		Email:         profissional.Email,
		Especialidade: profissional.Especialidade,
		Img:           profissional.Img,
	}, nil
}
