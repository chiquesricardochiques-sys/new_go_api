package cliente

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrDadosInvalidos = errors.New("dados inválidos")
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func validarIDs(req RelacaoRequest) error {
	if req.AuthUserID <= 0 || req.SalaoID <= 0 {
		return fmt.Errorf(
			"%w: auth_user_id e salao_id são obrigatórios",
			ErrDadosInvalidos,
		)
	}

	return nil
}

// normalizarTelefone deixa só os dígitos e garante o DDI do Brasil (55),
// para o número servir direto no link do WhatsApp (wa.me/55...).
//
//	"(19) 99999-8888" -> "5519999998888"
//	"+55 19 99999-8888" -> "5519999998888"
//	"" -> "" (telefone não informado, é permitido)
func normalizarTelefone(bruto string) (string, error) {
	var digitos strings.Builder

	for _, c := range bruto {
		if c >= '0' && c <= '9' {
			digitos.WriteRune(c)
		}
	}

	telefone := digitos.String()

	if telefone == "" {
		if strings.TrimSpace(bruto) != "" {
			return "", fmt.Errorf("%w: telefone inválido", ErrDadosInvalidos)
		}

		return "", nil
	}

	// 10 ou 11 dígitos = DDD + número, sem o código do país.
	if len(telefone) == 10 || len(telefone) == 11 {
		telefone = "55" + telefone
	}

	if len(telefone) != 12 && len(telefone) != 13 {
		return "", fmt.Errorf("%w: telefone inválido", ErrDadosInvalidos)
	}

	return telefone, nil
}

// Verificar somente consulta a relação existente.
func (s *Service) Verificar(
	ctx context.Context,
	req RelacaoRequest,
) (*RelacaoResponse, error) {
	if err := validarIDs(req); err != nil {
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

	relacao, err := s.repository.BuscarRelacao(
		ctx,
		req.AuthUserID,
	)

	if err != nil {
		return nil, err
	}

	return &RelacaoResponse{
		IDCliente: relacao.ID,
		Telefone:  relacao.Telefone,
	}, nil
}

// Adicionar cria a relação caso não exista.
// Se já existir, retorna o mesmo id_cliente e, se vier telefone,
// atualiza o telefone do vínculo.
func (s *Service) Adicionar(
	ctx context.Context,
	req RelacaoRequest,
) (*RelacaoResponse, bool, error) {
	if err := validarIDs(req); err != nil {
		return nil, false, err
	}

	telefone, err := normalizarTelefone(req.Telefone)
	if err != nil {
		return nil, false, err
	}

	// Primeiro valida o salão.
	if err := s.repository.VerificarSalao(
		ctx,
		req.SalaoID,
	); err != nil {
		return nil, false, err
	}

	_, criado, err := s.repository.AdicionarRelacao(
		ctx,
		req.AuthUserID,
		telefone,
	)

	if err != nil {
		return nil, false, err
	}

	// Lê de volta para devolver o telefone que ficou salvo
	// (inclusive o antigo, quando a requisição veio sem telefone).
	relacao, err := s.repository.BuscarRelacao(
		ctx,
		req.AuthUserID,
	)

	if err != nil {
		return nil, false, err
	}

	return &RelacaoResponse{
		IDCliente: relacao.ID,
		Telefone:  relacao.Telefone,
	}, criado, nil
}