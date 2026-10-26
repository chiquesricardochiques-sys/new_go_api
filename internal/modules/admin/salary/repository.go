package salary

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("profissional não encontrado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BuscarProfissional(
	ctx context.Context,
	salaoID int64,
	profissionalID int64,
) (ProfessionalSalary, error) {
	var p ProfessionalSalary

	err := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			nome,
			especialidade,
			foto_url,
			tipo_remuneracao,
			percentual_comissao,
			salario_fixo,
			frequencia_pagamento
		FROM profissionais
		WHERE salao_id = ?
		  AND id = ?
		  AND deleted_at IS NULL
		  AND status = 'ativo'
		LIMIT 1
	`,
		salaoID,
		profissionalID,
	).Scan(
		&p.ID,
		&p.Nome,
		&p.Especialidade,
		&p.FotoURL,
		&p.TipoRemuneracao,
		&p.PercentualComissao,
		&p.SalarioFixo,
		&p.FrequenciaPagamento,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return ProfessionalSalary{}, ErrNotFound
	}

	if err != nil {
		return ProfessionalSalary{}, err
	}

	return p, nil
}

type PeriodoFinanceiro struct {
	Qtd         int
	TotalGerado float64
}

func (r *Repository) BuscarPeriodo(
	ctx context.Context,
	salaoID int64,
	profissionalID int64,
	inicio time.Time,
	fim time.Time,
) (PeriodoFinanceiro, error) {
	var periodo PeriodoFinanceiro

	err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(DISTINCT a.id) AS qtd,
			COALESCE(
				SUM(spa.preco_unitario * spa.quantidade),
				0
			) AS total_gerado
		FROM agendamentos a
		LEFT JOIN servico_por_agendamento spa
			ON spa.agendamento_id = a.id
		WHERE a.salao_id = ?
		  AND a.profissional_id = ?
		  AND a.status = 'concluido'
		  AND a.inicio >= ?
		  AND a.inicio < ?
	`,
		salaoID,
		profissionalID,
		inicio,
		fim,
	).Scan(
		&periodo.Qtd,
		&periodo.TotalGerado,
	)

	if err != nil {
		return PeriodoFinanceiro{}, err
	}

	return periodo, nil
}