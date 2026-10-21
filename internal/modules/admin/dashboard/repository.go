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
	return &Repository{db: db}
}

// Conta os agendamentos de um salão em um intervalo.
// O início é inclusivo e o fim é exclusivo.
func (r *Repository) ContarAgendamentos(
	ctx context.Context,
	idSalao int64,
	inicio time.Time,
	fim time.Time,
	cancelados bool,
) (int64, error) {
	condicaoStatus := "status <> 'cancelado'"

	if cancelados {
		condicaoStatus = "status = 'cancelado'"
	}

	query := `
		SELECT COUNT(*)
		FROM agendamentos
		WHERE salao_id = ?
		  AND inicio >= ?
		  AND inicio < ?
		  AND ` + condicaoStatus

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

// Retorna os totais de hoje e dos últimos sete dias
// para cada profissional ativo do salão.
func (r *Repository) ContarPorProfissional(
	ctx context.Context,
	idSalao int64,
	hoje time.Time,
	amanha time.Time,
	inicioSemana time.Time,
	fimSemana time.Time,
) ([]ProfissionalResumo, error) {
	const query = `
		SELECT
			p.id,
			p.nome,

			COALESCE(
				SUM(
					CASE
						WHEN a.inicio >= ?
						 AND a.inicio < ?
						 AND a.status <> 'cancelado'
						THEN 1
						ELSE 0
					END
				),
				0
			) AS total_hoje,

			COALESCE(
				SUM(
					CASE
						WHEN a.inicio >= ?
						 AND a.inicio < ?
						 AND a.status <> 'cancelado'
						THEN 1
						ELSE 0
					END
				),
				0
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

		GROUP BY p.id, p.nome
		ORDER BY p.nome ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		hoje,
		amanha,
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