package saloes

import "errors"

var (
	ErrDadosInvalidos     = errors.New("dados inválidos")
	ErrSalaoNaoEncontrado = errors.New("salão não encontrado")
	ErrNaoPodeAvaliar     = errors.New("atendimento não encontrado, não concluído ou já avaliado")
)

type BuscarRequest struct {
	Busca      string   `json:"busca"`
	Cidade     string   `json:"cidade"`
	Estado     string   `json:"estado"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	AuthUserID int64    `json:"auth_user_id"`
	Limite     int      `json:"limite"`
	Offset     int      `json:"offset"`
}

type PorSlugRequest struct {
	Slug string `json:"slug"`
}

type AvaliarRequest struct {
	AuthUserID    int64  `json:"auth_user_id"`
	AgendamentoID int64  `json:"agendamento_id"`
	Nota          int    `json:"nota"`
	Comentario    string `json:"comentario"`
}

type SalaoItem struct {
	ID       int64   `json:"id"`
	Nome     string  `json:"nome"`
	Slug     string  `json:"slug"`
	Endereco *string `json:"endereco"`
	Cidade   *string `json:"cidade"`
	Estado   *string `json:"estado"`
	Whatsapp *string `json:"whatsapp"`
	LogoURL  *string `json:"logo_url"`
	ImgSalao *string `json:"img_salao"`

	DistanciaKm    *float64 `json:"distancia_km"`
	TemAgendamento bool     `json:"tem_agendamento"`

	// Vitrine. Aberto é nil quando o salão ainda não cadastrou horário.
	Aberto             *bool    `json:"aberto"`
	NotaMedia          *float64 `json:"nota_media"`
	TotalAvaliacoes    int      `json:"total_avaliacoes"`
	PrecoInicial       *float64 `json:"preco_inicial"`
	TotalProfissionais int      `json:"total_profissionais"`
	Especialidades     []string `json:"especialidades"`
}

type BuscarData struct {
	Saloes  []SalaoItem `json:"saloes"`
	Limite  int         `json:"limite"`
	Offset  int         `json:"offset"`
	TemMais bool        `json:"tem_mais"`
}

type Cidade struct {
	Cidade string `json:"cidade"`
	Estado string `json:"estado"`
	Total  int    `json:"total"`
}

// ---------- Detalhes do salão ----------

type ServicoItem struct {
	ID         int64   `json:"id"`
	Nome       string  `json:"nome"`
	Descricao  *string `json:"descricao"`
	Img        *string `json:"img"`
	Preco      float64 `json:"preco"`
	DuracaoMin int     `json:"duracao_minutos"`
}

type ProfissionalItem struct {
	ID            int64   `json:"id"`
	Nome          string  `json:"nome"`
	Especialidade *string `json:"especialidade"`
	FotoURL       *string `json:"foto_url"`
}

type HorarioItem struct {
	Dia        string  `json:"dia"`
	Abre       bool    `json:"abre"`
	Abertura   *string `json:"abertura"`
	Fechamento *string `json:"fechamento"`
}

type AvaliacaoItem struct {
	Nota       int     `json:"nota"`
	Comentario *string `json:"comentario"`
	Data       string  `json:"data"`
}

type DetalhesData struct {
	Salao         SalaoItem          `json:"salao"`
	Servicos      []ServicoItem      `json:"servicos"`
	Profissionais []ProfissionalItem `json:"profissionais"`
	Horarios      []HorarioItem      `json:"horarios"`
	Avaliacoes    []AvaliacaoItem    `json:"avaliacoes"`
}

// Filtro interno (já validado pelo service)
type filtro struct {
	Busca      string
	Cidade     string
	Estado     string
	Latitude   *float64
	Longitude  *float64
	AuthUserID int64
	Limite     int
	Offset     int

	// "agora" no fuso do Brasil, calculado no service
	Dia  string // seg, ter, qua...
	Hora string // 15:04:05
}