package codaces

import (
	"database/sql"
	"errors"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// ------------------------------------------------------------
// BUSCAR CÓDIGO DO SALÃO
// ------------------------------------------------------------

func (r *Repository) BuscarCodigo(
	salaoID uint64,
) (*CodigoResponse, error) {

	var response CodigoResponse
	var codigo sql.NullString

	err := r.db.QueryRow(`
		SELECT id, codigo_painel
		FROM saloes
		WHERE id = ?
	`, salaoID).Scan(
		&response.SalaoID,
		&codigo,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSalaoNaoEncontrado
		}

		return nil, err
	}

	if !codigo.Valid || strings.TrimSpace(codigo.String) == "" {
		return nil, ErrCodigoNaoCadastrado
	}

	response.Codigo = codigo.String

	return &response, nil
}

// ------------------------------------------------------------
// VERIFICAR SE O SALÃO EXISTE
// ------------------------------------------------------------

func (r *Repository) SalaoExiste(
	salaoID uint64,
) (bool, error) {

	var existe bool

	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM saloes
			WHERE id = ?
		)
	`, salaoID).Scan(&existe)

	if err != nil {
		return false, err
	}

	return existe, nil
}

// ------------------------------------------------------------
// VERIFICAR SE O CÓDIGO JÁ PERTENCE A OUTRO SALÃO
// ------------------------------------------------------------

func (r *Repository) CodigoExisteEmOutroSalao(
	codigo string,
	salaoID uint64,
) (bool, error) {

	var existe bool

	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM saloes
			WHERE codigo_painel = ?
			  AND id <> ?
		)
	`, codigo, salaoID).Scan(&existe)

	if err != nil {
		return false, err
	}

	return existe, nil
}

// ------------------------------------------------------------
// INSERIR CÓDIGO
// Só permite inserir quando codigo_painel estiver NULL
// ------------------------------------------------------------

func (r *Repository) InserirCodigo(
	salaoID uint64,
	codigo string,
) error {

	result, err := r.db.Exec(`
		UPDATE saloes
		SET codigo_painel = ?
		WHERE id = ?
		  AND codigo_painel IS NULL
	`, codigo, salaoID)

	if err != nil {
		if isDuplicateKey(err) {
			return ErrCodigoDuplicado
		}

		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected > 0 {
		return nil
	}

	// Nenhuma linha foi alterada.
	// Verifica se o salão existe.
	existe, err := r.SalaoExiste(salaoID)
	if err != nil {
		return err
	}

	if !existe {
		return ErrSalaoNaoEncontrado
	}

	// Se o salão já possui código, não permite inserir outro.
	_, err = r.BuscarCodigo(salaoID)

	if err == nil {
		return ErrSalaoJaPossuiCodigo
	}

	if errors.Is(err, ErrCodigoNaoCadastrado) {
		return ErrCodigoNaoCadastrado
	}

	return err
}

// ------------------------------------------------------------
// ALTERAR CÓDIGO EXISTENTE
// ------------------------------------------------------------

func (r *Repository) AlterarCodigo(
	salaoID uint64,
	codigo string,
) error {

	_, err := r.db.Exec(`
		UPDATE saloes
		SET codigo_painel = ?
		WHERE id = ?
		  AND codigo_painel IS NOT NULL
	`, codigo, salaoID)

	if err != nil {
		if isDuplicateKey(err) {
			return ErrCodigoDuplicado
		}

		return err
	}

	// Não utilizamos RowsAffected para validar a alteração.
	// Se o código enviado for igual ao código atual,
	// o MySQL pode retornar zero alterações e isso não é erro.

	return nil
}

// ------------------------------------------------------------
// IDENTIFICAR CÓDIGO DUPLICADO NO MYSQL
// ------------------------------------------------------------

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}

	mensagem := strings.ToLower(err.Error())

	return strings.Contains(mensagem, "1062") ||
		strings.Contains(mensagem, "duplicate entry") ||
		strings.Contains(mensagem, "duplicate key") ||
		strings.Contains(mensagem, "unique constraint")
}