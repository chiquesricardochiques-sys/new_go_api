package profissional

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AddRelations(
	ctx context.Context,
	req AddRelationsRequest,
) error {
	if req.IDSalao <= 0 ||
		req.AuthUserID <= 0 ||
		strings.TrimSpace(req.Nome) == "" {
		return ErrDadosInvalidos
	}

	var salaoExiste bool

	err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(
			SELECT 1 FROM saloes WHERE id = ?
		)`,
		req.IDSalao,
	).Scan(&salaoExiste)

	if err != nil {
		return err
	}

	if !salaoExiste {
		return ErrSalaoInvalido
	}

	var vinculoExiste bool

	err = r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM profissionais
			WHERE salao_id = ? AND auth_user_id = ?
		)`,
		req.IDSalao,
		req.AuthUserID,
	).Scan(&vinculoExiste)

	if err != nil {
		return err
	}

	if vinculoExiste {
		return ErrDuplicado
	}

	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO profissionais (
			salao_id,
			auth_user_id,
			nome,
			status
		) VALUES (?, ?, ?, 'aguardando_liberacao')`,
		req.IDSalao,
		req.AuthUserID,
		strings.TrimSpace(req.Nome),
	)

	if err != nil {
		// Protege contra uma duplicação concorrente.
		if strings.Contains(
			strings.ToLower(err.Error()),
			"duplicate entry",
		) {
			return ErrDuplicado
		}

		return err
	}

	return nil
}

func (r *Repository) Verify(
	ctx context.Context,
	req VerifyRequest,
) (VerifyResponse, error) {
	var response VerifyResponse

	response.IDSalao = req.IDSalao
	response.AuthUserID = req.AuthUserID

	if req.IDSalao <= 0 || req.AuthUserID <= 0 {
		return response, ErrDadosInvalidos
	}

	err := r.db.QueryRowContext(
		ctx,
		`SELECT status
		FROM profissionais
		WHERE salao_id = ? AND auth_user_id = ?
		LIMIT 1`,
		req.IDSalao,
		req.AuthUserID,
	).Scan(&response.Status)

	if errors.Is(err, sql.ErrNoRows) {
		response.Existe = false
		response.Ativo = false
		response.Status = ""
		return response, nil
	}

	if err != nil {
		return response, err
	}

	response.Existe = true
	response.Ativo = response.Status == "ativo"

	return response, nil
}