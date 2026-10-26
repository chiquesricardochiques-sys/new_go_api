package conf_prof

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) BuscarValor(
	ctx context.Context,
	profissionalID int64,
	chave string,
) ([]byte, error) {

	var valor []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT valor
		FROM configuracao_profissional
		WHERE profissional_id = ?
		  AND chave = ?
		LIMIT 1
	`,
		profissionalID,
		chave,
	).Scan(&valor)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return valor, nil
}

func (r *Repository) DefinirValor(
	ctx context.Context,
	profissionalID int64,
	chave string,
	valor any,
) error {

	jsonValor, err := json.Marshal(valor)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO configuracao_profissional (
			profissional_id,
			chave,
			valor
		)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			valor = ?
	`,
		profissionalID,
		chave,
		jsonValor,
		jsonValor,
	)

	return err
}