package dashboard

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// ============================================================
// AGENDAMENTOS
// ============================================================

func (r *Repository) ContarAgendamentos(
	ctx context.Context,
	idSalao int64,
	inicio time.Time,
	fim time.Time,
) (int64, error) {

	const query = `
		SELECT COUNT(*)
		FROM agendamentos
		WHERE salao_id = ?
		  AND inicio >= ?
		  AND inicio < ?
		  AND status <> 'cancelado'
	`

	var total int64

	err := r.db.QueryRowContext(
		ctx,
		query,
		idSalao,
		inicio,
		fim,
	).Scan(&total)

	if err != nil {
		return 0, err
	}

	return total, nil
}

// ============================================================
// FATURAMENTO
// ============================================================

func (r *Repository) CalcularFaturamento(
	ctx context.Context,
	idSalao int64,
	inicio time.Time,
	fim time.Time,
) (float64, error) {

	const query = `
		SELECT
			COALESCE(
				SUM(
					spa.preco_unitario * spa.quantidade
				),
				0
			)
		FROM agendamentos a
		INNER JOIN servico_por_agendamento spa
			ON spa.agendamento_id = a.id
		WHERE a.salao_id = ?
		  AND a.status = 'concluido'
		  AND a.inicio >= ?
		  AND a.inicio < ?
	`

	var faturamento float64

	err := r.db.QueryRowContext(
		ctx,
		query,
		idSalao,
		inicio,
		fim,
	).Scan(&faturamento)

	if err != nil {
		return 0, err
	}

	return faturamento, nil
}

// ============================================================
// FATURAMENTO POR MÊS
// ============================================================

func (r *Repository) FaturamentoPorMes(
	ctx context.Context,
	idSalao int64,
) ([]FaturamentoPorMes, error) {
	const query = `
		SELECT
			YEAR(a.inicio) AS ano,
			MONTH(a.inicio) AS mes,
			DATE_FORMAT(a.inicio, '%Y-%m') AS periodo,
			COALESCE(
				SUM(spa.preco_unitario * spa.quantidade),
				0
			) AS faturamento
		FROM agendamentos a
		INNER JOIN servico_por_agendamento spa
			ON spa.agendamento_id = a.id
		WHERE a.salao_id = ?
			AND a.status = 'concluido'
		GROUP BY
			YEAR(a.inicio),
			MONTH(a.inicio),
			DATE_FORMAT(a.inicio, '%Y-%m')
		ORDER BY
			YEAR(a.inicio),
			MONTH(a.inicio)
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		idSalao,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultado := make([]FaturamentoPorMes, 0)

	for rows.Next() {
		var item FaturamentoPorMes

		if err := rows.Scan(
			&item.Ano,
			&item.Mes,
			&item.Periodo,
			&item.Faturamento,
		); err != nil {
			return nil, err
		}

		resultado = append(resultado, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resultado, nil
}
// ============================================================
// PROFISSIONAIS
// ============================================================

func (r *Repository) ContarPorProfissional(
	ctx context.Context,
	idSalao int64,
	inicioHoje time.Time,
	fimHoje time.Time,
	inicioSemana time.Time,
	fimSemana time.Time,
) ([]ProfissionalResumo, error) {

	const query = `
		SELECT
			p.id,
			p.nome,
			COALESCE(p.especialidade, '') AS especialidade,
			COALESCE(p.foto_url, '') AS foto_url,

			COUNT(
				CASE
					WHEN a.inicio >= ?
					 AND a.inicio < ?
					 AND a.status <> 'cancelado'
					THEN 1
				END
			) AS total_hoje,

			COUNT(
				CASE
					WHEN a.inicio >= ?
					 AND a.inicio < ?
					 AND a.status <> 'cancelado'
					THEN 1
				END
			) AS total_semana

		FROM profissionais p

		LEFT JOIN agendamentos a
			ON a.profissional_id = p.id
			AND a.salao_id = p.salao_id
			AND a.inicio >= ?
			AND a.inicio < ?

		WHERE p.salao_id = ?
		  AND p.status = 'ativo'
		  AND p.deleted_at IS NULL

		GROUP BY
			p.id,
			p.nome,
			p.especialidade,
			p.foto_url

		ORDER BY p.nome ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		inicioHoje,
		fimHoje,
		inicioSemana,
		fimSemana,
		inicioSemana,
		fimSemana,
		idSalao,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	resultado := make([]ProfissionalResumo, 0)

	for rows.Next() {

		var item ProfissionalResumo

		err := rows.Scan(
			&item.ProfissionalID,
			&item.Nome,
			&item.Especialidade,
			&item.FotoURL,
			&item.TotalHoje,
			&item.TotalSemana,
		)

		if err != nil {
			return nil, err
		}

		resultado = append(resultado, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resultado, nil
}