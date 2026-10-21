package database

import (
	"context"
	"database/sql"
	"fmt"
)

func WithTransaction(
	ctx context.Context,
	db *sql.DB,
	fn func(*sql.Tx) error,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}

	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar transação: %w", err)
	}

	return nil
}
