package client_home

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const layoutDataHora = "2006-01-02 15:04:05"

var (
	ErrSalaoNaoEncontrado   = errors.New("salão não encontrado")
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

// BuscarProximoAgendamento devolve o primeiro agendamento ativo
// (pendente/confirmado) do cliente que ainda não terminou.
// "agora" vem do Go em horário de Brasília (não usa NOW() do MySQL,
// que pode estar em UTC). Retorna nil se não houver.
func (r *Repository) BuscarProximoAgendamento(
	ctx context.Context,
	salaoID int64,
	authUserID int64,
	agora time.Time,
) (*ProximoAgendamento, error) {
	const query = `
		SELECT
			a.id,
			DATE_FORMAT(a.inicio, '%Y-%m-%d %H:%i:%s'),
			DATE_FORMAT(a.fim,    '%Y-%m-%d %H:%i:%s'),
			a.status,
			p.nome
		FROM agendamentos a
		INNER JOIN profissionais p
			ON p.id = a.profissional_id
		WHERE a.salao_id = ?
			AND a.auth_user_id = ?
			AND a.status IN ('pendente', 'confirmado')
			AND a.fim > ?
		ORDER BY a.inicio ASC, a.id ASC
		LIMIT 1
	`

	var (
		agendamento ProximoAgendamento
		inicio      string
		fim         string
	)

	err := r.db.QueryRowContext(
		ctx,
		query,
		salaoID,
		authUserID,
		agora.Format(layoutDataHora),
	).Scan(
		&agendamento.ID,
		&inicio,
		&fim,
		&agendamento.Status,
		&agendamento.Profissional,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	agendamento.Data = inicio[:10]
	agendamento.Hora = inicio[11:16]
	agendamento.HoraFim = fim[11:16]
	agendamento.Servicos = []string{}

	const servicosQuery = `
		SELECT
			nome_servico_snapshot,
			preco_unitario,
			duracao_minutos,
			quantidade
		FROM servico_por_agendamento
		WHERE agendamento_id = ?
		ORDER BY id ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		servicosQuery,
		agendamento.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			nome       string
			preco      float64
			duracao    int
			quantidade int
		)

		if err := rows.Scan(
			&nome,
			&preco,
			&duracao,
			&quantidade,
		); err != nil {
			return nil, err
		}

		agendamento.Servicos = append(agendamento.Servicos, nome)
		agendamento.DuracaoTotal += duracao * quantidade
		agendamento.PrecoTotal += preco * float64(quantidade)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &agendamento, nil
}