package salary

import "time"

type Request struct {
	SalaoID        int64  `json:"salao_id"`
	ProfissionalID int64  `json:"profissional_id"`
	MesAno         string `json:"mes_ano"`
}

type Periodo struct {
	Qtd         int     `json:"qtd"`
	TotalGerado float64 `json:"totalGerado"`
	Comissao    float64 `json:"comissao"`
	SalarioFixo float64 `json:"salarioFixo"`
	ValorGanho  float64 `json:"valorGanho"`
}

type ResponseData struct {
	Nome                string  `json:"nome"`
	Especialidade       string  `json:"especialidade"`
	FotoURL             *string `json:"foto_url"`
	TipoRemuneracao     string  `json:"tipo_remuneracao"`
	PercentualComissao  float64 `json:"percentual_comissao"`
	SalarioFixo         float64 `json:"salario_fixo"`
	FrequenciaPagamento string  `json:"frequencia_pagamento"`

	Hoje      Periodo `json:"hoje"`
	Semana    Periodo `json:"semana"`
	MesAtual  Periodo `json:"mesAtual"`
	MesFiltro Periodo `json:"mesFiltro"`
}

type ProfessionalSalary struct {
	ID                  int64
	Nome                string
	Especialidade       string
	FotoURL             *string
	TipoRemuneracao     string
	PercentualComissao  float64
	SalarioFixo         float64
	FrequenciaPagamento string
}

type PeriodoFiltro struct {
	Inicio time.Time
	Fim    time.Time
}