package professional_schedule

import (
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) ProfissionalPertenceAoSalao(
	profissionalID uint64,
	salaoID uint64,
) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM profissionais
			WHERE id = ?
			  AND salao_id = ?
			  AND deleted_at IS NULL
		)
	`

	var exists bool

	err := r.db.QueryRow(
		query,
		profissionalID,
		salaoID,
	).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Repository) ListarGrade(
	salaoID uint64,
	profissionalID uint64,
) ([]ScheduleDay, error) {
	const query = `
		SELECT
			dia_semana,
			abre,
			COALESCE(
				TIME_FORMAT(abertura, '%H:%i:%s'),
				''
			) AS abertura,
			COALESCE(
				TIME_FORMAT(pausa_inicio, '%H:%i:%s'),
				''
			) AS pausa_inicio,
			COALESCE(
				TIME_FORMAT(pausa_fim, '%H:%i:%s'),
				''
			) AS pausa_fim,
			COALESCE(
				TIME_FORMAT(fechamento, '%H:%i:%s'),
				''
			) AS fechamento
		FROM profissional_grade_semanal
		WHERE salao_id = ?
		  AND profissional_id = ?
		ORDER BY
			FIELD(
				dia_semana,
				'seg',
				'ter',
				'qua',
				'qui',
				'sex',
				'sab',
				'dom'
			)
	`

	rows, err := r.db.Query(
		query,
		salaoID,
		profissionalID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	grade := make([]ScheduleDay, 0)

	for rows.Next() {
		var item ScheduleDay

		if err := rows.Scan(
			&item.DiaSemana,
			&item.Abre,
			&item.Abertura,
			&item.PausaInicio,
			&item.PausaFim,
			&item.Fechamento,
		); err != nil {
			return nil, err
		}

		grade = append(
			grade,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return grade, nil
}

// InicializarGrade cria os 7 dias da semana para o profissional.
//
// Todos começam fechados:
// abre = false
// horários = NULL
//
// A função é idempotente: se os dias já existirem,
// não cria registros duplicados.
func (r *Repository) InicializarGrade(
	tx *sql.Tx,
	salaoID uint64,
	profissionalID uint64,
) error {
	const query = `
		INSERT INTO profissional_grade_semanal (
			salao_id,
			profissional_id,
			dia_semana,
			abre,
			abertura,
			pausa_inicio,
			pausa_fim,
			fechamento
		)
		VALUES
			(?, ?, 'seg', 0, NULL, NULL, NULL, NULL),
			(?, ?, 'ter', 0, NULL, NULL, NULL, NULL),
			(?, ?, 'qua', 0, NULL, NULL, NULL, NULL),
			(?, ?, 'qui', 0, NULL, NULL, NULL, NULL),
			(?, ?, 'sex', 0, NULL, NULL, NULL, NULL),
			(?, ?, 'sab', 0, NULL, NULL, NULL, NULL),
			(?, ?, 'dom', 0, NULL, NULL, NULL, NULL)
		ON DUPLICATE KEY UPDATE
			id = id
	`

	_, err := tx.Exec(
		query,

		salaoID,
		profissionalID,

		salaoID,
		profissionalID,

		salaoID,
		profissionalID,

		salaoID,
		profissionalID,

		salaoID,
		profissionalID,

		salaoID,
		profissionalID,

		salaoID,
		profissionalID,
	)

	return err
}

func (r *Repository) SalvarGrade(
	tx *sql.Tx,
	salaoID uint64,
	profissionalID uint64,
	grade []ScheduleDay,
) error {
	const deleteQuery = `
		DELETE FROM profissional_grade_semanal
		WHERE salao_id = ?
		  AND profissional_id = ?
	`

	if _, err := tx.Exec(
		deleteQuery,
		salaoID,
		profissionalID,
	); err != nil {
		return err
	}

	if len(grade) == 0 {
		return nil
	}

	const insertQuery = `
		INSERT INTO profissional_grade_semanal (
			salao_id,
			profissional_id,
			dia_semana,
			abre,
			abertura,
			pausa_inicio,
			pausa_fim,
			fechamento
		)
		VALUES (
			?,
			?,
			?,
			?,
			NULLIF(?, ''),
			NULLIF(?, ''),
			NULLIF(?, ''),
			NULLIF(?, '')
		)
	`

	stmt, err := tx.Prepare(insertQuery)

	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, item := range grade {
		if _, err := stmt.Exec(
			salaoID,
			profissionalID,
			item.DiaSemana,
			item.Abre,
			item.Abertura,
			item.PausaInicio,
			item.PausaFim,
			item.Fechamento,
		); err != nil {
			return fmt.Errorf(
				"erro ao inserir grade de %s: %w",
				item.DiaSemana,
				err,
			)
		}
	}

	return nil
}

func (r *Repository) ResetarGrade(
	tx *sql.Tx,
	salaoID uint64,
	profissionalID uint64,
) error {
	const query = `
		DELETE FROM profissional_grade_semanal
		WHERE salao_id = ?
		  AND profissional_id = ?
	`

	_, err := tx.Exec(
		query,
		salaoID,
		profissionalID,
	)

	return err
}