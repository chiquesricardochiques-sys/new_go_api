package cliente

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrRelacaoNaoEncontrada = errors.New("relação do cliente não encontrada")
	ErrSalaoNaoEncontrado   = errors.New("salão não encontrado")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// Verifica se o salão existe e está ativo.
func (r *Repository) VerificarSalao(
	ctx context.Context,
	salaoID int64,
) error {
	const query = `
		SELECT id
		FROM saloes
		WHERE id = ?
			AND status = 'ativo'
		LIMIT 1
	`

	var id int64

	err := r.db.QueryRowContext(
		ctx,
		query,
		salaoID,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrSalaoNaoEncontrado
	}

	if err != nil {
		return err
	}

	return nil
}

// Busca o vínculo global entre Auth e cliente.
//
// O campo "id" da tabela clientes_auth_vinculos
// é o ID interno do cliente utilizado pela aplicação.
func (r *Repository) BuscarRelacao(
	ctx context.Context,
	authUserID int64,
) (int64, error) {
	const query = `
		SELECT id
		FROM clientes_auth_vinculos
		WHERE auth_user_id = ?
		LIMIT 1
	`

	var idCliente int64

	err := r.db.QueryRowContext(
		ctx,
		query,
		authUserID,
	).Scan(&idCliente)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrRelacaoNaoEncontrada
	}

	if err != nil {
		return 0, err
	}

	return idCliente, nil
}

// Adiciona o vínculo caso ainda não exista.
//
// A UNIQUE KEY em auth_user_id garante que o mesmo usuário
// não tenha dois vínculos.
//
// Se o vínculo já existir, o LAST_INSERT_ID(id) faz com que
// LastInsertId() retorne o ID existente.
func (r *Repository) AdicionarRelacao(
	ctx context.Context,
	authUserID int64,
) (int64, bool, error) {
	const query = `
		INSERT INTO clientes_auth_vinculos (
			auth_user_id
		)
		VALUES (?)
		ON DUPLICATE KEY UPDATE
			id = LAST_INSERT_ID(id)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		authUserID,
	)

	if err != nil {
		return 0, false, err
	}

	idCliente, err := result.LastInsertId()
	if err != nil {
		return 0, false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, false, err
	}

	criado := affected == 1

	return idCliente, criado, nil
}