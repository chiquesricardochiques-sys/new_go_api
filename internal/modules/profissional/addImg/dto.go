package addimg

import "errors"

// ============================================================================
// ERROS
// ============================================================================

var (
	ErrDadosInvalidos = errors.New("dados inválidos")
	ErrNaoEncontrado  = errors.New("profissional não encontrado")
	ErrJaPossuiFoto   = errors.New("o profissional já possui foto, use a alteração")
	ErrSemFoto        = errors.New("o profissional não possui foto")
)

// ============================================================================
// LIMITES (iguais às colunas da tabela profissionais)
// ============================================================================

const (
	MaxURL      = 500 // foto_url       varchar(500)
	MaxPublicID = 255 // foto_public_id varchar(255)
)

// ============================================================================
// REQUESTS (tudo vem do app)
// ============================================================================

// FotoRequest é usado para adicionar e alterar a foto.
type FotoRequest struct {
	IDSalao        int64  `json:"id_salao"`
	ProfissionalID int64  `json:"profissional_id"`
	FotoURL        string `json:"foto_url"`
	FotoPublicID   string `json:"foto_public_id"`
}

// RemoverRequest é usado para limpar a foto.
type RemoverRequest struct {
	IDSalao        int64 `json:"id_salao"`
	ProfissionalID int64 `json:"profissional_id"`
}

// ============================================================================
// ENTIDADE / RESPONSE
// ============================================================================

type Foto struct {
	URL      string `json:"foto_url"`
	PublicID string `json:"foto_public_id"`
}

func (f Foto) Existe() bool {
	return f.URL != "" || f.PublicID != ""
}

// FotoResponse:
//   - foto     = o que ficou salvo agora
//   - anterior = a foto que foi substituída/removida (o app usa o public_id
//     dela para apagar a imagem antiga na API de imagens). Vem vazio se não
//     havia foto antes ou se o public_id continuou o mesmo.
type FotoResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
	Foto     *Foto  `json:"foto,omitempty"`
	Anterior *Foto  `json:"anterior,omitempty"`
}