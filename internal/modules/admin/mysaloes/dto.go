
package mysaloes

import "errors"

var (
	ErrImagemNaoEncontrada = errors.New("imagem não encontrada ou indisponível")
	ErrDadosInvalidos     = errors.New("dados inválidos")
	ErrSemAcesso          = errors.New("salão não encontrado ou sem acesso")
	ErrSalaoIndisponivel  = errors.New("este salão não está disponível para acesso")
	ErrEdicaoIndisponivel = errors.New("este salão não pode ser editado no status atual")
	ErrLimitePendentes    = errors.New("você já tem salões aguardando liberação; aguarde a análise antes de criar outro")
	ErrSlugIndisponivel   = errors.New("não foi possível gerar um endereço único para o salão")

	ErrLimiteImagens = errors.New("este salão já possui o limite de 3 imagens")
)

const (
	StatusAtivo     = "ativo"
	StatusPendente  = "pendente"
	StatusRejeitado = "rejeitado"
)

// ---------- Requests ----------

type ListarRequest struct {
	AuthUserID int64 `json:"auth_user_id"`
}

type CriarRequest struct {
	AuthUserID int64 `json:"auth_user_id"`

	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
	Telefone  string `json:"telefone"`
	Email     string `json:"email"`
	Whatsapp  string `json:"whatsapp"`

	SiteURL             string `json:"site_url"`
	Instagram           string `json:"instagram"`
	Facebook            string `json:"facebook"`
	TipoEstabelecimento string `json:"tipo_estabelecimento"`
	Documento           string `json:"documento"`

	ChavePix       string `json:"chave_pix"`
	ChavePixTipo   string `json:"chave_pix_tipo"`
	NomeTitularPix string `json:"nome_titular_pix"`

	Endereco string `json:"endereco"`
	Cidade   string `json:"cidade"`
	Estado   string `json:"estado"`
	CEP      string `json:"cep"`

	// Imagens de apresentação armazenadas na tabela saloes.
	LogoURL  string `json:"logo_url"`
	ImgSalao string `json:"img_salao"`
}

// ---------- Dados completos para edição do salão ----------

type DadosAtualizacaoSalao struct {
	Salao   SalaoResumo   `json:"salao"`
	Imagens []ImagemSalao `json:"imagens"`
}

type AtualizarRequest struct {
	CriarRequest

	IDSalao int64 `json:"id_salao"`
}

type EntrarRequest struct {
	AuthUserID int64 `json:"auth_user_id"`
	IDSalao    int64 `json:"id_salao"`
}

// ---------- Imagens da galeria do salão ----------

type AdicionarImagemRequest struct {
	AuthUserID int64  `json:"auth_user_id"`
	IDSalao    int64  `json:"id_salao"`
	URL        string `json:"url"`
	PublicID   string `json:"public_id"`
	Descricao  string `json:"descricao"`
	Ordem      int    `json:"ordem"`
	Principal  bool   `json:"principal"`
}


// ---------- Atualização de imagem da galeria ----------

type AtualizarImagemRequest struct {
	AuthUserID int64  `json:"auth_user_id"`
	IDSalao    int64  `json:"id_salao"`
	IDImagem   int64  `json:"id_imagem"`
	URL        string `json:"url"`
	PublicID   string `json:"public_id"`
}

type ListarImagensRequest struct {
	AuthUserID int64 `json:"auth_user_id"`
	IDSalao    int64 `json:"id_salao"`
}

type ImagemSalao struct {
	ID        int64   `json:"id"`
	IDSalao   int64   `json:"id_salao"`
	URL       string  `json:"url"`
	PublicID  string  `json:"public_id"`
	Descricao *string `json:"descricao"`
	Ordem     int     `json:"ordem"`
	Principal bool    `json:"principal"`
}

// ---------- Responses ----------

type SalaoResumo struct {
	ID           int64   `json:"id"`
	Nome         string  `json:"nome"`
	Descricao    *string `json:"descricao"`
	Slug         string  `json:"slug"`
	CodigoPainel *string `json:"codigo_painel"`

	Telefone *string `json:"telefone"`
	Email    *string `json:"email"`
	Whatsapp *string `json:"whatsapp"`

	SiteURL             *string `json:"site_url"`
	Instagram           *string `json:"instagram"`
	Facebook            *string `json:"facebook"`
	TipoEstabelecimento *string `json:"tipo_estabelecimento"`
	Documento           *string `json:"documento"`

	ChavePix       *string `json:"chave_pix"`
	ChavePixTipo   *string `json:"chave_pix_tipo"`
	NomeTitularPix *string `json:"nome_titular_pix"`

	Endereco *string `json:"endereco"`
	Cidade   *string `json:"cidade"`
	Estado   *string `json:"estado"`
	CEP      *string `json:"cep"`

	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`

	LogoURL  *string `json:"logo_url"`
	ImgSalao *string `json:"img_salao"`

	Status        string  `json:"status"`
	MotivoAnalise *string `json:"motivo_analise"`
	Papel         string  `json:"papel"`

	TemAvisosNaoLidos int     `json:"tem_avisos_nao_lidos"`
	ImagemCard        *string `json:"imagem_card"`
}

// ---------- Dados validados pelo service ----------

type NovoSalao struct {
	Nome      string
	Descricao string
	Telefone  string
	Email     string
	Whatsapp  string

	SiteURL             string
	Instagram           string
	Facebook            string
	TipoEstabelecimento string
	Documento           string

	ChavePix       string
	ChavePixTipo   string
	NomeTitularPix string

	Endereco string
	Cidade   string
	Estado   string
	CEP      string

	// Imagens de apresentação armazenadas na tabela saloes.
	LogoURL  string
	ImgSalao string
}
