package services

import "time"

type Servico struct {
	ID             uint64    `json:"id"`
	Nome           string    `json:"nome"`
	Descricao      *string   `json:"descricao"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ServicoSalao struct {
	ID              uint64    `json:"id"`
	SalaoServicoID  uint64    `json:"salao_servico_id"`
	Nome            string    `json:"nome"`
	Descricao       *string   `json:"descricao"`
	Preco           float64   `json:"preco"`
	DuracaoMinutos  uint      `json:"duracao_minutos"`
	Status          string    `json:"status"`
	Selecionado     bool      `json:"selecionado"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ServicoProfissionalItem struct {
	ID              uint64  `json:"id"`
	SalaoServicoID  uint64  `json:"salao_servico_id"`
	Nome            string  `json:"nome"`
	Descricao       *string `json:"descricao"`
	Preco           float64 `json:"preco"`
	DuracaoMinutos  uint    `json:"duracao_minutos"`
	Status          string  `json:"status"`
	Selecionado     bool    `json:"selecionado"`
}

type Profissional struct {
	ID            uint64  `json:"id"`
	Nome          string  `json:"nome"`
	Especialidade *string `json:"especialidade"`
	Status        string  `json:"status"`
}

type ProfissionalServico struct {
	ID             uint64    `json:"id"`
	SalaoID        uint64    `json:"salao_id"`
	ProfissionalID uint64    `json:"profissional_id"`
	ServicoID      uint64    `json:"servico_id"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type ConfiguracaoServicoSalao struct {
	ServicoID      uint64  `json:"servico_id"`
	Preco          float64 `json:"preco"`
	DuracaoMinutos uint    `json:"duracao_minutos"`
}