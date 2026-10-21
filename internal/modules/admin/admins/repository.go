
package admins

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNaoEncontrado = errors.New("vínculo administrativo não encontrado")
	ErrDuplicado     = errors.New("esta conta já possui um vínculo com o salão")
	ErrSalaoInvalido = errors.New("salão não encontrado")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AddRelations(
	ctx context.Context,
	idSalao int64,
	authUserID int64,
) error {
	var existeSalao bool

	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM saloes WHERE id = ?
		)`,
		idSalao,
	).Scan(&existeSalao)
	if err != nil {
		return err
	}

	if !existeSalao {
		return ErrSalaoInvalido
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO salao_admins (
			salao_id,
			auth_user_id,
			status
		)
		SELECT ?, ?, 'ativo'
		WHERE NOT EXISTS (
			SELECT 1
			FROM salao_admins
			WHERE salao_id = ?
			  AND auth_user_id = ?
		)
	`, idSalao, authUserID, idSalao, authUserID)
	if err != nil {
		return err
	}

	afetadas, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if afetadas == 0 {
		return ErrDuplicado
	}

	return nil
}

func (r *Repository) Verify(
	ctx context.Context,
	idSalao int64,
	authUserID int64,
) (VerifyResponse, error) {
	var resposta VerifyResponse

	resposta.IDSalao = idSalao
	resposta.AuthUserID = authUserID

	err := r.db.QueryRowContext(ctx, `
		SELECT status
		FROM salao_admins
		WHERE salao_id = ?
		  AND auth_user_id = ?
		LIMIT 1
	`, idSalao, authUserID).Scan(&resposta.Status)

	if errors.Is(err, sql.ErrNoRows) {
		resposta.Existe = false
		resposta.Ativo = false
		return resposta, nil
	}

	if err != nil {
		return VerifyResponse{}, err
	}

	resposta.Existe = true
	resposta.Ativo = resposta.Status == "ativo"

	return resposta, nil
}
