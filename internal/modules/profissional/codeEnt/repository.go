package codeEnt

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

func (r *Repository) BuscarPorCodigo(
	ctx context.Context,
	codigo string,
) (*SalaoResponse, error) {

	var salao SalaoResponse

	err := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			nome,
			slug,
			status
		FROM saloes
		WHERE codigo_painel = ?
		LIMIT 1
	`, codigo).Scan(
		&salao.ID,
		&salao.Nome,
		&salao.Slug,
		&salao.Status,
	)

	if err == sql.ErrNoRows {
		return nil, ErrSalaoNaoEncontrado
	}

	if err != nil {
		return nil, err
	}

	return &salao, nil
}