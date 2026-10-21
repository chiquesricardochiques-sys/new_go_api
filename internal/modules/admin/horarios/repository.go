package horarios

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("horário não encontrado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Listar(
	ctx context.Context,
	idSalao int64,
	profissionalID int64,
	data string,
) ([]Horario, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, profissional_id, data, hora, status
		FROM horarios
		WHERE id_salao = ?
		  AND profissional_id = ?
		  AND data = ?
		ORDER BY hora ASC
	`, idSalao, profissionalID, data)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultado := make([]Horario, 0)

	for rows.Next() {
		var h Horario
		var dataDB, horaDB []byte

		if err := rows.Scan(
			&h.ID,
			&h.ProfissionalID,
			&dataDB,
			&horaDB,
			&h.Status,
		); err != nil {
			return nil, err
		}

		h.Data = string(dataDB)
		h.Hora = string(horaDB)

		resultado = append(resultado, h)
	}

	return resultado, rows.Err()
}

func (r *Repository) Criar(
	ctx context.Context,
	idSalao int64,
	profissionalID int64,
	data string,
	hora string,
) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO horarios
			(id_salao, profissional_id, data, hora, status)
		VALUES (?, ?, ?, ?, 'livre')
	`, idSalao, profissionalID, data, hora)

	return err
}

func (r *Repository) Excluir(
	ctx context.Context,
	idSalao int64,
	id int64,
) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM horarios
		WHERE id = ? AND id_salao = ?
	`, id, idSalao)
	if err != nil {
		return err
	}

	afetados, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if afetados == 0 {
		return ErrNotFound
	}

	return nil
}
