package profissional

import "errors"

var (
	ErrDadosInvalidos     = errors.New("dados inválidos")
	ErrNaoEncontrado      = errors.New("registro não encontrado")
	ErrProfissionalInativo = errors.New("este profissional não está ativo")
)

// ============================================================================
// REQUESTS
// ============================================================================

// BaseRequest é o que quase toda rota recebe: antes era project_id + instance_id,
// agora é só id_salao.
type BaseRequest struct {
	IDSalao        int64 `json:"id_salao"`
	ProfissionalID int64 `json:"profissional_id"`
}

// MeRequest substitui o antigo login-profissional: o login agora é feito no Auth,
// aqui só resolvemos "quem é o profissional desse auth_user_id neste salão".
type MeRequest struct {
	IDSalao    int64 `json:"id_salao"`
	AuthUserID int64 `json:"auth_user_id"`
}

type AgendaRequest struct {
	BaseRequest
	Data   string `json:"data"`   // YYYY-MM-DD (vazio = hoje)
	Status string `json:"status"` // vazio ou "todos" = sem filtro
}

type AlterarStatusRequest struct {
	BaseRequest
	AgendamentoID int64  `json:"agendamento_id"`
	NovoStatus    string `json:"novo_status"`
}

type HistoricoRequest struct {
	BaseRequest
	Periodo string `json:"periodo"` // dia | semana | mes | vazio (tudo)
}

type HorariosDiaRequest struct {
	BaseRequest
	Data string `json:"data"` // YYYY-MM-DD
}

type FaturamentoRequest struct {
	BaseRequest
	Ano int `json:"ano"`
	Mes int `json:"mes"`
}

// ============================================================================
// ENTIDADES / RESPONSES
// ============================================================================

type Profissional struct {
	ID                  int64   `json:"id"`
	IDSalao             int64   `json:"id_salao"`
	AuthUserID          *int64  `json:"auth_user_id"`
	Nome                string  `json:"nome"`
	Especialidade       string  `json:"especialidade"`
	Img                 string  `json:"img"`
	Status              string  `json:"status"`
	Ativo               int     `json:"ativo"`
	CriadoEm            string  `json:"criado_em"`
	TipoRemuneracao     string  `json:"tipo_remuneracao"`
	PercentualComissao  float64 `json:"percentual_comissao"`
	SalarioFixo         float64 `json:"salario_fixo"`
	FrequenciaPagamento string  `json:"frequencia_pagamento"`
	UsaHorarios         int     `json:"usa_horarios"`
}

type PerfilResponse struct {
	Profissional Profissional `json:"profissional"`
}

type AgendamentoItem struct {
	ID                 int64    `json:"id"`
	Data               string   `json:"data"`
	Hora               string   `json:"hora"`
	HoraFim            string   `json:"hora_fim"`
	UsaHorarios        int      `json:"usa_horarios"`
	Status             string   `json:"status"`
	Observacoes        string   `json:"observacoes"`
	ClienteID          int64    `json:"cliente_id"`
	Cliente            string   `json:"cliente"`
	ClienteTelefone    *string  `json:"cliente_telefone"`
	WhatsappLink       *string  `json:"whatsapp_link"`
	Servicos           []string `json:"servicos"`
	DuracaoTotal       int      `json:"duracao_total"`
	PrecoTotal         float64  `json:"preco_total"`
	PercentualComissao float64  `json:"percentual_comissao"`
}

type AgendaResponse struct {
	Agendamentos []AgendamentoItem `json:"agendamentos"`
	Data         string            `json:"data"`
}

type AlterarStatusResponse struct {
	NovoStatus string `json:"novo_status"`
}

type ResumoResponse struct {
	TotalHoje   int `json:"totalHoje"`
	TotalSemana int `json:"totalSemana"`
	TotalMes    int `json:"totalMes"`
}

type HistoricoResponse struct {
	TotalConcluidos int               `json:"totalConcluidos"`
	TotalCancelados int               `json:"totalCancelados"`
	Historico       []AgendamentoItem `json:"historico"`
}

type Slot struct {
	Hora        string   `json:"hora"`
	Disponivel  bool     `json:"disponivel"`
	Agendamento *int64   `json:"agendamento"`
	Servicos    []string `json:"servicos"`
}

type HorariosDiaResponse struct {
	UsaHorarios int    `json:"usa_horarios"`
	Bloqueado   bool   `json:"bloqueado"`
	Horarios    []Slot `json:"horarios"`
}

type FaturamentoProfissional struct {
	Nome                string  `json:"nome"`
	TipoRemuneracao     string  `json:"tipo_remuneracao"`
	FrequenciaPagamento string  `json:"frequencia_pagamento"`
	PercentualComissao  float64 `json:"percentual_comissao"`
	SalarioFixo         float64 `json:"salario_fixo"`
}

type FaturamentoPeriodo struct {
	Ano            int    `json:"ano"`
	Mes            int    `json:"mes"`
	QuantidadeDias int    `json:"quantidade_dias"`
	DataInicio     string `json:"data_inicio"`
	DataFim        string `json:"data_fim"`
}

type FaturamentoEstatisticas struct {
	TotalAtendimentos int     `json:"total_atendimentos"`
	FaturamentoTotal  float64 `json:"faturamento_total"`
	ValorComissao     float64 `json:"valor_comissao"`
	ValorFixo         float64 `json:"valor_fixo"`
	ValorAReceber     float64 `json:"valor_a_receber"`
}

type FaturamentoResponse struct {
	Profissional FaturamentoProfissional `json:"profissional"`
	Periodo      FaturamentoPeriodo      `json:"periodo"`
	Estatisticas FaturamentoEstatisticas `json:"estatisticas"`
}

// ============================================================================
// TIPOS INTERNOS (repository <-> service)
// ============================================================================

type Grade struct {
	Abre        bool
	Abertura    string // HH:MM
	PausaInicio string // HH:MM ou ""
	PausaFim    string // HH:MM ou ""
	Fechamento  string // HH:MM
}

type FiltroAgendamentos struct {
	IDSalao        int64
	ProfissionalID int64
	Inicio         string // "YYYY-MM-DD HH:MM:SS" (inclusivo) ou "" = sem limite
	Fim            string // "YYYY-MM-DD HH:MM:SS" (exclusivo) ou "" = sem limite
	Status         string // "" = todos
	Desc           bool
}

type LinhaFaturamento struct {
	ID                 int64
	PrecoTotal         float64
	PercentualComissao float64
}

type BloqueioHorario struct {
	Inicio string // "YYYY-MM-DD HH:MM"
	Fim    string
}

// ClienteInfo / ClienteLookup: o nome e o telefone do cliente moram no Auth,
// não neste banco. Quem implementar essa interface (chamando o Auth) preenche
// cliente/whatsapp na agenda. Se for nil, a agenda devolve só o cliente_id.
type ClienteInfo struct {
	Nome     string
	Telefone string
}