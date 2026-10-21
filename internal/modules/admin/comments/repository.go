package comments

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("comentário não encontrado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Listar(ctx context.Context, idSalao int64) ([]Comment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, nome, comentario, COALESCE(img, ''),
		       COALESCE(ativo, 0)
		FROM comments
		WHERE id_salao = ?
		ORDER BY id DESC
	`, idSalao)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comentarios := make([]Comment, 0)

	for rows.Next() {
		var c Comment
		var ativo int

		if err := rows.Scan(
			&c.ID,
			&c.Nome,
			&c.Comentario,
			&c.Img,
			&ativo,
		); err != nil {
			return nil, err
		}

		c.Ativo = ativo == 1
		comentarios = append(comentarios, c)
	}

	return comentarios, rows.Err()
}

func (r *Repository) Criar(
	ctx context.Context,
	idSalao int64,
	nome string,
	comentario string,
	imagem CommentImage,
) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO comments
			(id_salao, nome, comentario, img, img_public_id, ativo)
		VALUES (?, ?, ?, ?, ?, 1)
	`,
		idSalao,
		nome,
		comentario,
		imagem.URL,
		nullString(imagem.PublicID),
	)

	return err
}

func (r *Repository) Buscar(
	ctx context.Context,
	idSalao int64,
	id int64,
) (*Comment, string, error) {
	var c Comment
	var ativo int
	var publicID sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, nome, comentario, COALESCE(img, ''),
		       COALESCE(ativo, 0), img_public_id
		FROM comments
		WHERE id = ? AND id_salao = ?
		LIMIT 1
	`, id, idSalao).Scan(
		&c.ID,
		&c.Nome,
		&c.Comentario,
		&c.Img,
		&ativo,
		&publicID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", ErrNotFound
	}
	if err != nil {
		return nil, "", err
	}

	c.Ativo = ativo == 1

	return &c, publicID.String, nil
}

func (r *Repository) Atualizar(
	ctx context.Context,
	idSalao int64,
	id int64,
	nome string,
	comentario string,
	ativo bool,
	imagem *CommentImage,
) error {
	var result sql.Result
	var err error

	ativoInt := 0
	if ativo {
		ativoInt = 1
	}

	if imagem != nil {
		result, err = r.db.ExecContext(ctx, `
			UPDATE comments
			SET nome = ?, comentario = ?, ativo = ?,
			    img = ?, img_public_id = ?
			WHERE id = ? AND id_salao = ?
		`,
			nome,
			comentario,
			ativoInt,
			imagem.URL,
			nullString(imagem.PublicID),
			id,
			idSalao,
		)
	} else {
		result, err = r.db.ExecContext(ctx, `
			UPDATE comments
			SET nome = ?, comentario = ?, ativo = ?
			WHERE id = ? AND id_salao = ?
		`,
			nome,
			comentario,
			ativoInt,
			id,
			idSalao,
		)
	}

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		_, _, err := r.Buscar(ctx, idSalao, id)
		return err
	}

	return nil
}

func (r *Repository) Excluir(
	ctx context.Context,
	idSalao int64,
	id int64,
) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM comments
		WHERE id = ? AND id_salao = ?
	`, id, idSalao)
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

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
