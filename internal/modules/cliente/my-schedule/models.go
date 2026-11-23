package myschedule

import "time"

const (
	CenarioMeusAgendamentos = "meus_agendamentos"
	CenarioHistorico        = "historico"
)

// ============================================================
// REQUESTS
// ============================================================

type ListarRequest struct {
	AuthUserID int64  `json:"auth_user_id"`
	Cenario    string `json:"cenario"`
	Status    string `json:"status"`
}

type CancelarRequest struct {
	AuthUserID    int64 `json:"auth_user_id"`
	AgendamentoID int64 `json:"agendamento_id"`
}

// ============================================================
// RESPONSES
// ============================================================

type ServicoAgendado struct {
	Nome       string  `json:"nome"`
	Preco      float64 `json:"preco"`
	DuracaoMin int     `json:"duracao_min"`
	Quantidade  int     `json:"quantidade"`
}

type Agendamento struct {
	ID           int64              `json:"id"`
	Salao        string             `json:"salao"`
	Data         string             `json:"data"`
	Hora         string             `json:"hora"`
	HoraFim      string             `json:"hora_fim"`
	Status       string             `json:"status"`
	Observacoes  *string            `json:"observacoes"`
	Profissional string             `json:"profissional"`
	Servicos     []ServicoAgendado `json:"servicos"`
	DuracaoTotal int                `json:"duracao_total"`
	PrecoTotal   float64            `json:"preco_total"`
	PodeCancelar bool               `json:"pode_cancelar"`

	Inicio time.Time `json:"-"`
}

type ListarData struct {
	Textos       map[string]string `json:"textos"`
	Cenario      string            `json:"cenario"`
	Status       *string           `json:"status"`
	Agendamentos []Agendamento     `json:"agendamentos"`
}

type ListarResponse struct {
	Success bool        `json:"success"`
	Data    *ListarData `json:"data,omitempty"`
}

type CancelarResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}