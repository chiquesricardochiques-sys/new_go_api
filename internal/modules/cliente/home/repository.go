package home

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrSalaoNaoEncontrado = errors.New("salão não encontrado")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) BuscarTextos(
	ctx context.Context,
	salaoID int64,
) (map[string]string, error) {
	const query = `
		SELECT
			chave,
			conteudo
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

	textos := make(map[string]string)

	for rows.Next() {
		var (
			chave    string
			conteudo string
		)

		if err := rows.Scan(
			&chave,
			&conteudo,
		); err != nil {
			return nil, err
		}

		textos[chave] = conteudo
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return textos, nil
}

func (r *Repository) BuscarSalao(
	ctx context.Context,
	salaoID int64,
) (*Salao, error) {
	const query = `
		SELECT
			id,
			nome,
			telefone,
			email,
			whatsapp,
			endereco,
			cidade,
			estado,
			cep,
			latitude,
			longitude,
			logo_url
		FROM saloes
		WHERE id = ?
			AND status = 'ativo'
		LIMIT 1
	`

	var salao Salao

	err := r.db.QueryRowContext(
		ctx,
		query,
		salaoID,
	).Scan(
		&salao.ID,
		&salao.Nome,
		&salao.Telefone,
		&salao.Email,
		&salao.Whatsapp,
		&salao.Endereco,
		&salao.Cidade,
		&salao.Estado,
		&salao.CEP,
		&salao.Latitude,
		&salao.Longitude,
		&salao.LogoURL,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSalaoNaoEncontrado
	}

	if err != nil {
		return nil, err
	}

	return &salao, nil
}

func (r *Repository) BuscarProfissionais(
	ctx context.Context,
	salaoID int64,
) ([]Profissional, error) {
	const query = `
		SELECT
			id,
			nome,
			especialidade,
			foto_url
		FROM profissionais
		WHERE salao_id = ?
			AND status = 'ativo'
			AND deleted_at IS NULL
		ORDER BY nome ASC
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

	profissionais := make([]Profissional, 0)

	for rows.Next() {
		var profissional Profissional

		if err := rows.Scan(
			&profissional.ID,
			&profissional.Nome,
			&profissional.Especialidade,
			&profissional.FotoURL,
		); err != nil {
			return nil, err
		}

		profissionais = append(
			profissionais,
			profissional,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profissionais, nil
}