package blocks

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrProfissionalNaoEncontrado = errors.New("profissional não encontrado neste salão")
	ErrBloqueioNaoEncontrado     = errors.New("bloqueio não encontrado")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ProfissionalExiste(
	ctx context.Context,
	idSalao, profissionalID int64,
) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM salao_beleza_profissionais
			WHERE id = ? AND id_salao = ?
			  AND deleted_at IS NULL
		)
	`

	var existe bool
	err := r.db.QueryRowContext(
		ctx, query, profissionalID, idSalao,
	).Scan(&existe)

	return existe, err
}

func (r *Repository) ListarDias(
	ctx context.Context,
	idSalao, profissionalID int64,
	hoje time.Time,
) ([]DiaBloqueado, error) {
	const query = `
		SELECT id, profissional_id, data, motivo
		FROM salao_beleza_dia_null
		WHERE id_salao = ?
		  AND profissional_id = ?
		  AND data >= ?
		ORDER BY data ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		idSalao,
		profissionalID,
		hoje.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultado := make([]DiaBloqueado, 0)

	for rows.Next() {
		var item DiaBloqueado
		var data time.Time
		var motivo sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.ProfissionalID,
			&data,
			&motivo,
		); err != nil {
			return nil, err
		}

		item.Data = data.Format("2006-01-02")
		if motivo.Valid {
			item.Motivo = motivo.String
		}

		resultado = append(resultado, item)
	}

	return resultado, rows.Err()
}

func (r *Repository) ListarHoras(
	ctx context.Context,
	idSalao, profissionalID int64,
	dataSelecionada string,
) ([]HoraBloqueada, error) {
	const query = `
		SELECT id, professional_id, data, hora, motivo
		FROM salao_beleza_orarios_null
		WHERE id_salao = ?
		  AND professional_id = ?
		  AND data = ?
		ORDER BY hora ASC
	`

	rows, err := r.db.QueryContext(
		ctx, query, idSalao, profissionalID, dataSelecionada,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultado := make([]HoraBloqueada, 0)

	for rows.Next() {
		var item HoraBloqueada
		var data time.Time
		var hora time.Time
		var motivo sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.ProfissionalID,
			&data,
			&hora,
			&motivo,
		); err != nil {
			return nil, err
		}

		item.Data = data.Format("2006-01-02")
		item.Hora = hora.Format("15:04:05")
		if motivo.Valid {
			item.Motivo = motivo.String
		}

		resultado = append(resultado, item)
	}

	return resultado, rows.Err()
}

func (r *Repository) AdicionarDia(
	ctx context.Context,
	idSalao, profissionalID int64,
	data, motivo string,
) error {
	const query = `
		INSERT INTO salao_beleza_dia_null
			(id_salao, profissional_id, data, motivo)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx, query, idSalao, profissionalID, data, nullableString(motivo),
	)
	return err
}

func (r *Repository) AdicionarHora(
	ctx context.Context,
	idSalao, profissionalID int64,
	data, hora, motivo string,
) error {
	const query = `
		INSERT INTO salao_beleza_orarios_null
			(id_salao, professional_id, data, hora, motivo)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		idSalao,
		profissionalID,
		data,
		hora,
		nullableString(motivo),
	)
	return err
}

func (r *Repository) RemoverDia(
	ctx context.Context,
	idSalao, id int64,
) error {
	return remover(ctx, r.db, `
		DELETE FROM salao_beleza_dia_null
		WHERE id = ? AND id_salao = ?
	`, id, idSalao)
}

func (r *Repository) RemoverHora(
	ctx context.Context,
	idSalao, id int64,
) error {
	return remover(ctx, r.db, `
		DELETE FROM salao_beleza_orarios_null
		WHERE id = ? AND id_salao = ?
	`, id, idSalao)
}

func remover(
	ctx context.Context,
	db *sql.DB,
	query string,
	id, idSalao int64,
) error {
	result, err := db.ExecContext(ctx, query, id, idSalao)
	if err != nil {
		return err
	}

	afetadas, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if afetadas == 0 {
		return ErrBloqueioNaoEncontrado
	}

	return nil
}

func nullableString(valor string) any {
	if valor == "" {
		return nil
	}
	return valor
}
