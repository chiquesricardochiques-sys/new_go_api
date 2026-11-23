package admins

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

// Verify diz se a conta tem vínculo com o salão e se ele está ativo.
// O vínculo agora é criado em mysaloes (ao criar o salão), nunca automaticamente.
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