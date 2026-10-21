package textos_salao

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var ErrNotFound = errors.New(
	"texto do salão não encontrado",
)

var ErrDuplicate = errors.New(
	"já existe um texto com essa chave para este salão",
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

const selectFields = `
	id,
	salao_id,
	chave,
	conteudo,
	updated_at
`

func scanTexto(
	scanner interface {
		Scan(dest ...any) error
	},
) (TextoSalao, error) {
	var texto TextoSalao

	err := scanner.Scan(
		&texto.ID,
		&texto.SalaoID,
		&texto.Chave,
		&texto.Conteudo,
		&texto.UpdatedAt,
	)

	if err != nil {
		return TextoSalao{}, err
	}

	return texto, nil
}

func (r *Repository) Listar(
	ctx context.Context,
	salaoID int64,
) ([]TextoSalao, error) {
	query := `
		SELECT ` + selectFields + `
		FROM textos_salao
		WHERE salao_id = ?
		ORDER BY chave ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		salaoID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	textos := make(
		[]TextoSalao,
		0,
	)

	for rows.Next() {
		texto, err := scanTexto(rows)

		if err != nil {
			return nil, err
		}

		textos = append(
			textos,
			texto,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return textos, nil
}

func (r *Repository) Buscar(
	ctx context.Context,
	salaoID int64,
	chave string,
) (TextoSalao, error) {
	query := `
		SELECT ` + selectFields + `
		FROM textos_salao
		WHERE salao_id = ?
		  AND chave = ?
		LIMIT 1
	`

	texto, err := scanTexto(
		r.db.QueryRowContext(
			ctx,
			query,
			salaoID,
			chave,
		),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return TextoSalao{}, ErrNotFound
	}

	if err != nil {
		return TextoSalao{}, err
	}

	return texto, nil
}

func (r *Repository) Adicionar(
	ctx context.Context,
	req *AddRequest,
) (TextoSalao, error) {
	query := `
		INSERT INTO textos_salao (
			salao_id,
			chave,
			conteudo
		)
		VALUES (?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		req.SalaoID,
		req.Chave,
		req.Conteudo,
	)

	if err != nil {
		if strings.Contains(
			err.Error(),
			"Duplicate entry",
		) {
			return TextoSalao{}, ErrDuplicate
		}

		return TextoSalao{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return TextoSalao{}, err
	}

	return r.BuscarPorID(
		ctx,
		id,
	)
}

func (r *Repository) BuscarPorID(
	ctx context.Context,
	id int64,
) (TextoSalao, error) {
	query := `
		SELECT ` + selectFields + `
		FROM textos_salao
		WHERE id = ?
		LIMIT 1
	`

	texto, err := scanTexto(
		r.db.QueryRowContext(
			ctx,
			query,
			id,
		),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return TextoSalao{}, ErrNotFound
	}

	if err != nil {
		return TextoSalao{}, err
	}

	return texto, nil
}

func (r *Repository) Atualizar(
	ctx context.Context,
	req *UpdateRequest,
) (TextoSalao, error) {
	query := `
		UPDATE textos_salao
		SET
			conteudo = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE salao_id = ?
		  AND chave = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		req.Conteudo,
		req.SalaoID,
		req.Chave,
	)

	if err != nil {
		return TextoSalao{}, err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		return TextoSalao{}, err
	}

	if affected == 0 {
		return TextoSalao{}, ErrNotFound
	}

	return r.Buscar(
		ctx,
		req.SalaoID,
		req.Chave,
	)
}

func (r *Repository) Excluir(
	ctx context.Context,
	req *DeleteRequest,
) error {
	query := `
		DELETE FROM textos_salao
		WHERE salao_id = ?
		  AND chave = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		req.SalaoID,
		req.Chave,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrNotFound
	}

	return nil
}