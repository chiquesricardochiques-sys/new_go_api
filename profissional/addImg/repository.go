package addimg

import (
	"context"
	"database/sql"
	"errors"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ============================================================================
// FOTO DO PROFISSIONAL
// Sempre filtra por salao_id + id: nunca lê nem altera profissional de outro salão.
// ============================================================================

// BuscarFoto devolve a foto atual (campos vazios se não tiver).
func (r *Repository) BuscarFoto(ctx context.Context, salaoID, profissionalID int64) (*Foto, error) {
	var url, publicID sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT foto_url, foto_public_id
		FROM profissionais
		WHERE id = ? AND salao_id = ? AND deleted_at IS NULL
		LIMIT 1`,
		profissionalID, salaoID,
	).Scan(&url, &publicID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNaoEncontrado
	}
	if err != nil {
		return nil, err
	}

	return &Foto{URL: url.String, PublicID: publicID.String}, nil
}

// SalvarFoto grava foto_url + foto_public_id (adicionar e alterar).
func (r *Repository) SalvarFoto(ctx context.Context, salaoID, profissionalID int64, foto Foto) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE profissionais
		SET foto_url = ?, foto_public_id = ?
		WHERE id = ? AND salao_id = ? AND deleted_at IS NULL`,
		foto.URL, foto.PublicID, profissionalID, salaoID,
	)
	return err
}

// RemoverFoto limpa foto_url e foto_public_id.
func (r *Repository) RemoverFoto(ctx context.Context, salaoID, profissionalID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE profissionais
		SET foto_url = NULL, foto_public_id = NULL
		WHERE id = ? AND salao_id = ? AND deleted_at IS NULL`,
		profissionalID, salaoID,
	)
	return err
}