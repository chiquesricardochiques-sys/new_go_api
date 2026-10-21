package salao

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("salão não encontrado")

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
	nome,
	slug,
	telefone,
	email,
	whatsapp,
	endereco,
	cidade,
	estado,
	cep,
	latitude,
	longitude,
	logo_url,
	status,
	created_at,
	updated_at
`

func scanSalao(scanner interface {
	Scan(dest ...any) error
}) (Salao, error) {
	var s Salao

	err := scanner.Scan(
		&s.ID,
		&s.Nome,
		&s.Slug,
		&s.Telefone,
		&s.Email,
		&s.Whatsapp,
		&s.Endereco,
		&s.Cidade,
		&s.Estado,
		&s.CEP,
		&s.Latitude,
		&s.Longitude,
		&s.LogoURL,
		&s.Status,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		return s, err
	}

	return s, nil
}

func (r *Repository) Buscar(
	ctx context.Context,
	salaoID int64,
) (Salao, error) {
	query := `
		SELECT ` + selectFields + `
		FROM saloes
		WHERE id = ?
		LIMIT 1
	`

	salao, err := scanSalao(
		r.db.QueryRowContext(
			ctx,
			query,
			salaoID,
		),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Salao{}, ErrNotFound
	}

	if err != nil {
		return Salao{}, err
	}

	return salao, nil
}

func (r *Repository) Atualizar(
	ctx context.Context,
	req *UpdateRequest,
) error {
	current, err := r.Buscar(
		ctx,
		req.SalaoID,
	)

	if err != nil {
		return err
	}

	if req.Nome != nil {
		current.Nome = *req.Nome
	}

	if req.Slug != nil {
		current.Slug = *req.Slug
	}

	if req.Telefone != nil {
		current.Telefone = normalizarStringNula(*req.Telefone)
	}

	if req.Email != nil {
		current.Email = normalizarStringNula(*req.Email)
	}

	if req.Whatsapp != nil {
		current.Whatsapp = normalizarStringNula(*req.Whatsapp)
	}

	if req.Endereco != nil {
		current.Endereco = normalizarStringNula(*req.Endereco)
	}

	if req.Cidade != nil {
		current.Cidade = normalizarStringNula(*req.Cidade)
	}

	if req.Estado != nil {
		current.Estado = normalizarStringNula(*req.Estado)
	}

	if req.CEP != nil {
		current.CEP = normalizarStringNula(*req.CEP)
	}

	if req.Latitude != nil {
		current.Latitude = req.Latitude
	}

	if req.Longitude != nil {
		current.Longitude = req.Longitude
	}

	if req.LogoURL != nil {
		current.LogoURL = normalizarStringNula(*req.LogoURL)
	}

	if req.Status != nil {
		current.Status = *req.Status
	}

	query := `
		UPDATE saloes
		SET
			nome = ?,
			slug = ?,
			telefone = ?,
			email = ?,
			whatsapp = ?,
			endereco = ?,
			cidade = ?,
			estado = ?,
			cep = ?,
			latitude = ?,
			longitude = ?,
			logo_url = ?,
			status = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		current.Nome,
		current.Slug,
		current.Telefone,
		current.Email,
		current.Whatsapp,
		current.Endereco,
		current.Cidade,
		current.Estado,
		current.CEP,
		current.Latitude,
		current.Longitude,
		current.LogoURL,
		current.Status,
		req.SalaoID,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		_, err = r.Buscar(
			ctx,
			req.SalaoID,
		)

		return err
	}

	return nil
}

func normalizarStringNula(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}