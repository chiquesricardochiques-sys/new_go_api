package painel

import "errors"

var (
	ErrDadosInvalidos = errors.New("dados inválidos")
	ErrCodigoInvalido = errors.New("código do salão não encontrado")
	ErrDataInvalida   = errors.New("data inválida")
)

// ============================================================
// REQUEST
// ============================================================

// AgendaRequest: o código identifica o salão (aceita "#0815332").
// data vazia = hoje (horário de Brasília).
// profissional_id = 0 mostra todos.
type AgendaRequest struct {
	Codigo         string `json:"codigo"`
	Data           string `json:"data"`
	ProfissionalID int64  `json:"profissional_id"`
}

// ============================================================
// RESPONSE
// ============================================================

type SalaoPainel struct {
	ID      int64   `json:"id"`
	Nome    string  `json:"nome"`
	Cidade  *string `json:"cidade"`
	Estado  *string `json:"estado"`
	LogoURL *string `json:"logo_url"`
}

type ProfissionalPainel struct {
	ID            int64   `json:"id"`
	Nome          string  `json:"nome"`
	Especialidade *string `json:"especialidade"`
	FotoURL       *string `json:"foto_url"`
}

// ClientePainel: origem "app" = cliente que agendou pelo site/app;
// "avulso" = agendamento feito pelo próprio salão.
type ClientePainel struct {
	Nome     string `json:"nome"`
	Telefone string `json:"telefone"`
	Origem   string `json:"origem"`
}

type ServicoPainel struct {
	Nome       string  `json:"nome"`
	Quantidade int     `json:"quantidade"`
	Preco      float64 `json:"preco"`
	DuracaoMin int     `json:"duracao_min"`
}

type AgendamentoPainel struct {
	ID             int64           `json:"id"`
	Hora           string          `json:"hora"`     // HH:MM
	HoraFim        string          `json:"hora_fim"` // HH:MM
	DuracaoMinutos int             `json:"duracao_minutos"`
	Status         string          `json:"status"`
	StatusLabel    string          `json:"status_label"`
	ProfissionalID int64           `json:"profissional_id"`
	Profissional   string          `json:"profissional"`
	Cliente        ClientePainel   `json:"cliente"`
	Servicos       []ServicoPainel `json:"servicos"`
	PrecoTotal     float64         `json:"preco_total"`
	Observacao     string          `json:"observacao"`
}

type ResumoDia struct {
	Total         int     `json:"total"`
	Pendentes     int     `json:"pendentes"`
	Confirmados   int     `json:"confirmados"`
	Concluidos    int     `json:"concluidos"`
	Cancelados    int     `json:"cancelados"`
	NaoCompareceu int     `json:"nao_compareceu"`
	// soma de pendentes + confirmados + concluídos
	ValorPrevisto float64 `json:"valor_previsto"`
}

type AgendaData struct {
	Salao          SalaoPainel          `json:"salao"`
	Profissionais  []ProfissionalPainel `json:"profissionais"`
	Data           string               `json:"data"`       // YYYY-MM-DD
	DiaSemana      string               `json:"dia_semana"` // "sábado"
	Hoje           bool                 `json:"hoje"`
	ProfissionalID int64                `json:"profissional_id"`
	Resumo         ResumoDia            `json:"resumo"`
	Agendamentos   []AgendamentoPainel  `json:"agendamentos"`
}

type AgendaResponse struct {
	Success bool        `json:"success"`
	Data    *AgendaData `json:"data,omitempty"`
}