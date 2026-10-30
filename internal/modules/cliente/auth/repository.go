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

// Relacao é o que existe em clientes_auth_vinculos para um usuário do Auth.
type Relacao struct {
	ID       int64
	Telefone string // vazio quando ainda não foi informado
}

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
) (*Relacao, error) {
	const query = `
		SELECT id, telefone
		FROM clientes_auth_vinculos
		WHERE auth_user_id = ?
		LIMIT 1
	`

	var (
		relacao  Relacao
		telefone sql.NullString
	)

	err := r.db.QueryRowContext(
		ctx,
		query,
		authUserID,
	).Scan(&relacao.ID, &telefone)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRelacaoNaoEncontrada
	}

	if err != nil {
		return nil, err
	}

	relacao.Telefone = telefone.String

	return &relacao, nil
}

// Adiciona o vínculo caso ainda não exista.
//
// A UNIQUE KEY em auth_user_id garante que o mesmo usuário
// não tenha dois vínculos.
//
// Telefone:
//   - vazio ("") não grava nada e, se o vínculo já existe,
//     mantém o telefone que já estava salvo;
//   - preenchido grava no vínculo novo ou atualiza o do vínculo existente.
//
// Se o vínculo já existir, o LAST_INSERT_ID(id) faz com que
// LastInsertId() retorne o ID existente.
//
// RowsAffected: 1 = inseriu, 2 = já existia e o telefone mudou,
// 0 = já existia e nada mudou. Por isso "criado" continua sendo affected == 1.
func (r *Repository) AdicionarRelacao(
	ctx context.Context,
	authUserID int64,
	telefone string,
) (int64, bool, error) {
	const query = `
		INSERT INTO clientes_auth_vinculos (
			auth_user_id,
			telefone
		)
		VALUES (?, NULLIF(?, ''))
		ON DUPLICATE KEY UPDATE
			id = LAST_INSERT_ID(id),
			telefone = COALESCE(NULLIF(?, ''), telefone)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		authUserID,
		telefone,
		telefone,
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