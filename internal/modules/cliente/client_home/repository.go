package client_home

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrSalaoNaoEncontrado = errors.New("salão não encontrado")
	ErrClienteNaoEncontrado = errors.New("cliente não encontrado")
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

func (r *Repository) BuscarAuthUserID(
	ctx context.Context,
	idCliente int64,
) (int64, error) {
	const query = `
		SELECT
			auth_user_id
		FROM clientes_auth_vinculos
		WHERE id = ?
		LIMIT 1
	`

	var authUserID int64

	err := r.db.QueryRowContext(
		ctx,
		query,
		idCliente,
	).Scan(&authUserID)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrClienteNaoEncontrado
	}

	if err != nil {
		return 0, err
	}

	return authUserID, nil
}
func (r *Repository) BuscarProximoAgendamento(
    ctx context.Context,
    salaoID int64,
    idCliente int64,
) (*ProximoAgendamento, error) {

    authUserID, err := r.BuscarAuthUserID(ctx, idCliente)
    if err != nil {
        return nil, err
    }

    const query = `
        SELECT
            a.id,
            a.hora_marcada,
            a.duracao_minutos,
            a.status,
            p.nome AS profissional
        FROM agendamentos a
        INNER JOIN profissionais p
            ON p.id = a.profissional_id
        WHERE a.salao_id = ?
          AND a.auth_user_id = ?
          AND a.status IN ('pendente', 'confirmado')
          AND a.hora_marcada >= NOW()
        ORDER BY a.hora_marcada ASC
        LIMIT 1
    `

    var agendamento ProximoAgendamento
    var horaMarcada time.Time
    var duracaoMinutos int

    err = r.db.QueryRowContext(
        ctx,
        query,
        salaoID,
        authUserID,
    ).Scan(
        &agendamento.ID,
        &horaMarcada,
        &duracaoMinutos,
        &agendamento.Status,
        &agendamento.Profissional,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }

        return nil, err
    }

    agendamento.Inicio = horaMarcada
    agendamento.Fim = horaMarcada.Add(
        time.Duration(duracaoMinutos) * time.Minute,
    )

    return &agendamento, nil
}