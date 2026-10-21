package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var ErrUsuarioNaoEncontrado = errors.New("usuário não encontrado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type Admin struct {
	ID        int64
	IDSalao   int64
	AuthUserID int64
	Nome      string
	Email     string
	SenhaHash string
	CriadoEm  string
}

type Profissional struct {
	ID            int64
	IDSalao       int64
	Nome          string
	Email         string
	SenhaHash     string
	Especialidade string
	Img           string
	Ativo         bool
}

// BuscarAdmin procura o vínculo do usuário com o salão.
// A senha deve ser validada pelo serviço AUTH externo.
func (r *Repository) BuscarAdmin(
	ctx context.Context,
	idSalao int64,
	email string,
) (*Admin, error) {
	// A tabela salao_admins não contém e-mail ou senha.
	// Por isso, o ID do usuário precisa ser resolvido pelo AUTH
	// antes de consultar o vínculo administrativo.
	return nil, fmt.Errorf(
		"BuscarAdmin requer o auth_user_id obtido pelo AUTH externo; salão=%d, email=%q",
		idSalao,
		email,
	)
}

// BuscarAdminPorAuthUserID verifica o vínculo administrativo
// usando o ID do usuário validado pelo AUTH externo.
func (r *Repository) BuscarAdminPorAuthUserID(
	ctx context.Context,
	idSalao int64,
	authUserID int64,
) (*Admin, error) {
	const query = `
		SELECT
			sa.id,
			sa.salao_id,
			sa.auth_user_id,
			s.nome,
			COALESCE(s.email, ''),
			sa.created_at
		FROM salao_admins sa
		INNER JOIN saloes s
			ON s.id = sa.salao_id
		WHERE sa.salao_id = ?
			AND sa.auth_user_id = ?
			AND sa.status = 'ativo'
			AND s.status = 'ativo'
		LIMIT 1
	`

	var admin Admin
	var criadoEm time.Time

	err := r.db.QueryRowContext(
		ctx,
		query,
		idSalao,
		authUserID,
	).Scan(
		&admin.ID,
		&admin.IDSalao,
		&admin.AuthUserID,
		&admin.Nome,
		&admin.Email,
		&criadoEm,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUsuarioNaoEncontrado
	}
	if err != nil {
		return nil, err
	}

	admin.CriadoEm = criadoEm.Format(time.RFC3339)
	return &admin, nil
}

// BuscarProfissional consulta a tabela atual de profissionais.
// A autenticação e a senha devem ser verificadas pelo AUTH externo.
func (r *Repository) BuscarProfissional(
	ctx context.Context,
	idSalao int64,
	email string,
) (*Profissional, error) {
	// A tabela profissionais fornecida não possui email nem senha_hash.
	// É necessário obter o auth_user_id do AUTH e relacioná-lo
	// ao profissional por meio do vínculo definido no seu sistema.
	return nil, fmt.Errorf(
		"BuscarProfissional requer o vínculo do usuário com profissionais; salão=%d, email=%q",
		idSalao,
		email,
	)
}

// BuscarProfissionalPorID consulta um profissional pelo ID
// e pelo salão ao qual pertence.
func (r *Repository) BuscarProfissionalPorID(
	ctx context.Context,
	idSalao int64,
	profissionalID int64,
) (*Profissional, error) {
	const query = `
		SELECT
			id,
			salao_id,
			nome,
			COALESCE(especialidade, ''),
			COALESCE(foto_url, ''),
			status
		FROM profissionais
		WHERE salao_id = ?
			AND id = ?
			AND status = 'ativo'
			AND deleted_at IS NULL
		LIMIT 1
	`

	var profissional Profissional
	var status string

	err := r.db.QueryRowContext(
		ctx,
		query,
		idSalao,
		profissionalID,
	).Scan(
		&profissional.ID,
		&profissional.IDSalao,
		&profissional.Nome,
		&profissional.Especialidade,
		&profissional.Img,
		&status,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUsuarioNaoEncontrado
	}
	if err != nil {
		return nil, err
	}

	profissional.Ativo = status == "ativo"
	return &profissional, nil
}

// ConverterIDAuth converte um ID recebido do AUTH externo.
func ConverterIDAuth(id string) (int64, error) {
	idConvertido, err := strconv.ParseInt(id, 10, 64)
	if err != nil || idConvertido <= 0 {
		return 0, errors.New("ID do usuário AUTH inválido")
	}

	return idConvertido, nil
}
