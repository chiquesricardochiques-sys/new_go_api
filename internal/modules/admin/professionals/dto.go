
package professionals

import "time"

type ListRequest struct {
	SalaoID int64 `json:"salao_id"`
	ID      int64 `json:"id,omitempty"`
}

type UpdateRequest struct {
	SalaoID             int64    `json:"salao_id"`
	ID                  int64    `json:"id"`
	Nome                *string  `json:"nome"`
	Especialidade       *string  `json:"especialidade"`
	Status              *string  `json:"status"`
	UsaHorarios         *bool    `json:"usa_horarios"`
	TipoRemuneracao     *string  `json:"tipo_remuneracao"`
	PercentualComissao  *float64 `json:"percentual_comissao"`
	SalarioFixo         *float64 `json:"salario_fixo"`
	FrequenciaPagamento *string  `json:"frequencia_pagamento"`
}

type Professional struct {
	ID                  int64     `json:"id"`
	SalaoID             int64     `json:"salao_id"`
	Nome                string    `json:"nome"`
	Especialidade       *string   `json:"especialidade"`
	FotoURL             *string   `json:"foto_url"`
	FotoPublicID        *string   `json:"foto_public_id"`
	Status              string    `json:"status"`
	UsaHorarios         bool      `json:"usa_horarios"`
	TipoRemuneracao     string    `json:"tipo_remuneracao"`
	PercentualComissao  float64   `json:"percentual_comissao"`
	SalarioFixo         float64   `json:"salario_fixo"`
	FrequenciaPagamento string    `json:"frequencia_pagamento"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
