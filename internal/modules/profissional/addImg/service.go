package addimg

import (
	"context"
	"fmt"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ============================================================================
// VALIDAÇÃO
// ============================================================================

// normalizar valida o que veio do app e devolve a Foto pronta para gravar.
//
// O public_id gerado pela API de imagens tem o formato
// {project_code}/{id_salao}/profissionais/{arquivo}; exigir esse trecho evita
// que alguém registre a imagem de outro salão ou de outra tabela.
func normalizar(idSalao, profissionalID int64, urlBruta, publicIDBruto string) (Foto, error) {
	if idSalao <= 0 || profissionalID <= 0 {
		return Foto{}, fmt.Errorf("%w: id_salao e profissional_id são obrigatórios", ErrDadosInvalidos)
	}

	url := strings.TrimSpace(urlBruta)
	publicID := strings.TrimSpace(publicIDBruto)

	if url == "" || publicID == "" {
		return Foto{}, fmt.Errorf("%w: foto_url e foto_public_id são obrigatórios", ErrDadosInvalidos)
	}
	if len(url) > MaxURL || len(publicID) > MaxPublicID {
		return Foto{}, fmt.Errorf("%w: foto_url ou foto_public_id passam do tamanho permitido", ErrDadosInvalidos)
	}
	if !strings.HasPrefix(url, "https://") {
		return Foto{}, fmt.Errorf("%w: foto_url deve começar com https://", ErrDadosInvalidos)
	}
	if !strings.Contains(publicID, fmt.Sprintf("/%d/profissionais/", idSalao)) {
		return Foto{}, fmt.Errorf("%w: foto_public_id não pertence aos profissionais deste salão", ErrDadosInvalidos)
	}
	if !strings.Contains(url, publicID) {
		return Foto{}, fmt.Errorf("%w: foto_url não corresponde ao foto_public_id", ErrDadosInvalidos)
	}

	return Foto{URL: url, PublicID: publicID}, nil
}

// ============================================================================
// ADICIONAR (primeira foto)
// ============================================================================

func (s *Service) Adicionar(ctx context.Context, req FotoRequest) (*Foto, error) {
	nova, err := normalizar(req.IDSalao, req.ProfissionalID, req.FotoURL, req.FotoPublicID)
	if err != nil {
		return nil, err
	}

	atual, err := s.repo.BuscarFoto(ctx, req.IDSalao, req.ProfissionalID)
	if err != nil {
		return nil, err
	}
	if atual.Existe() {
		return nil, ErrJaPossuiFoto
	}

	if err := s.repo.SalvarFoto(ctx, req.IDSalao, req.ProfissionalID, nova); err != nil {
		return nil, err
	}

	return &nova, nil
}

// ============================================================================
// ATUALIZAR (trocar a foto)
// Devolve também a foto anterior, para o app apagar a imagem antiga.
// ============================================================================

func (s *Service) Atualizar(ctx context.Context, req FotoRequest) (nova *Foto, anterior *Foto, err error) {
	foto, err := normalizar(req.IDSalao, req.ProfissionalID, req.FotoURL, req.FotoPublicID)
	if err != nil {
		return nil, nil, err
	}

	atual, err := s.repo.BuscarFoto(ctx, req.IDSalao, req.ProfissionalID)
	if err != nil {
		return nil, nil, err
	}

	if err := s.repo.SalvarFoto(ctx, req.IDSalao, req.ProfissionalID, foto); err != nil {
		return nil, nil, err
	}

	// Se o public_id é o mesmo, não há imagem antiga para apagar
	// (apagar seria apagar a que acabou de ser salva).
	if atual.Existe() && atual.PublicID != foto.PublicID {
		anterior = atual
	}

	return &foto, anterior, nil
}

// ============================================================================
// REMOVER
// Devolve a foto removida, para o app apagar a imagem na API de imagens.
// ============================================================================

func (s *Service) Remover(ctx context.Context, req RemoverRequest) (*Foto, error) {
	if req.IDSalao <= 0 || req.ProfissionalID <= 0 {
		return nil, fmt.Errorf("%w: id_salao e profissional_id são obrigatórios", ErrDadosInvalidos)
	}

	atual, err := s.repo.BuscarFoto(ctx, req.IDSalao, req.ProfissionalID)
	if err != nil {
		return nil, err
	}
	if !atual.Existe() {
		return nil, ErrSemFoto
	}

	if err := s.repo.RemoverFoto(ctx, req.IDSalao, req.ProfissionalID); err != nil {
		return nil, err
	}

	return atual, nil
}