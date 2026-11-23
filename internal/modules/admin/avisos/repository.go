package avisos

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Listar(
	ctx context.Context,
	salaoID int64,
	historico bool,
) ([]Aviso, error) {

	avisos := make([]Aviso, 0)

	// HISTÓRICO: busca todos os avisos sem alterar o estado.
	if historico {
		query := `
			SELECT
				id,
				salao_id,
				titulo,
				mensagem,
				tipo,
				lido,
				created_at
			FROM avisos_salao
			WHERE salao_id = ?
			ORDER BY created_at DESC, id DESC
		`

		rows, err := r.db.QueryContext(ctx, query, salaoID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var aviso Aviso

			if err := rows.Scan(
				&aviso.ID,
				&aviso.SalaoID,
				&aviso.Titulo,
				&aviso.Mensagem,
				&aviso.Tipo,
				&aviso.Lido,
				&aviso.CreatedAt,
			); err != nil {
				return nil, err
			}

			avisos = append(avisos, aviso)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		return avisos, nil
	}

	// AVISOS NOVOS:
	// Busca os avisos lido = 1 e depois marca como visualizados.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT
			id,
			salao_id,
			titulo,
			mensagem,
			tipo,
			lido,
			created_at
		FROM avisos_salao
		WHERE salao_id = ?
		  AND lido = 1
		ORDER BY created_at DESC, id DESC
		FOR UPDATE
	`

	rows, err := tx.QueryContext(ctx, query, salaoID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var aviso Aviso

		if err := rows.Scan(
			&aviso.ID,
			&aviso.SalaoID,
			&aviso.Titulo,
			&aviso.Mensagem,
			&aviso.Tipo,
			&aviso.Lido,
			&aviso.CreatedAt,
		); err != nil {
			rows.Close()
			return nil, err
		}

		avisos = append(avisos, aviso)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	// Marca como visualizados somente os avisos que foram buscados.
	_, err = tx.ExecContext(
		ctx,
		`
			UPDATE avisos_salao
			SET lido = 0
			WHERE salao_id = ?
			  AND lido = 1
		`,
		salaoID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return avisos, nil
}