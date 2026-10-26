package dadosagendamentos

import "errors"

var (
	ErrDadosInvalidos = errors.New("dados inválidos")
	ErrNaoEncontrado  = errors.New("agendamento não encontrado")
	ErrNaoPendente    = errors.New("só é possível ver os detalhes de agendamentos pendentes")
)

type DadosRequest struct {
	SalaoID       int64 `json:"salao_id"`
	AgendamentoID int64 `json:"agendamento_id"`
	IDCliente     int64 `json:"id_cliente"`
}

type ProfissionalDados struct {
	ID           int64  `json:"id"`
	Nome         string `json:"nome"`
	Especialidade string `json:"especialidade"`
	FotoURL      string `json:"foto_url"`
}

type SalaoDados struct {
	Nome     string `json:"nome"`
	Telefone string `json:"telefone"`
	Whatsapp string `json:"whatsapp"`
	Endereco string `json:"endereco"`
	Cidade   string `json:"cidade"`
	Estado   string `json:"estado"`
}

type ServicoDados struct {
	ID             int64   `json:"id"`
	Nome           string  `json:"nome"`
	FotoURL        string  `json:"foto_url"`
	Quantidade     int     `json:"quantidade"`
	DuracaoMinutos int     `json:"duracao_minutos"`
	PrecoUnitario  float64 `json:"preco_unitario"`
	Subtotal       float64 `json:"subtotal"`
}

type AgendamentoDados struct {
	ID             int64             `json:"id"`
	Status         string            `json:"status"`
	StatusLabel    string            `json:"status_label"`
	Data           string            `json:"data"`
	DiaSemana      string            `json:"dia_semana"`
	Hora           string            `json:"hora"`
	HoraFim        string            `json:"hora_fim"`
	DuracaoMinutos int               `json:"duracao_minutos"`
	Profissional   ProfissionalDados `json:"profissional"`
	Salao          SalaoDados        `json:"salao"`
	Servicos       []ServicoDados    `json:"servicos"`
	PrecoTotal     float64           `json:"preco_total"`
	Observacao     string            `json:"observacao"`
	CriadoEm       string            `json:"criado_em"`
}

type DadosResponse struct {
	Success bool               `json:"success"`
	Data    *AgendamentoDados  `json:"data,omitempty"`
}