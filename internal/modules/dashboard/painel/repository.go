package painel

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

// BuscarSalaoPorCodigo localiza o salão ativo pelo código do painel.
func (r *Repository) BuscarSalaoPorCodigo(
	ctx context.Context,
	codigo string,
) (*SalaoPainel, error) {
	const query = `
		SELECT id, nome, cidade, estado, logo_url
		FROM saloes
		WHERE codigo_painel = ?
		  AND status = 'ativo'
		LIMIT 1
	`

	var salao SalaoPainel

	err := r.db.QueryRowContext(ctx, query, codigo).Scan(
		&salao.ID,
		&salao.Nome,
		&salao.Cidade,
		&salao.Estado,
		&salao.LogoURL,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCodigoInvalido
	}

	if err != nil {
		return nil, err
	}

	return &salao, nil
}

func (r *Repository) BuscarProfissionais(
	ctx context.Context,
	salaoID int64,
) ([]ProfissionalPainel, error) {
	const query = `
		SELECT id, nome, especialidade, foto_url
		FROM profissionais
		WHERE salao_id = ?
		  AND status = 'ativo'
		  AND deleted_at IS NULL
		ORDER BY nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query, salaoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profissionais := make([]ProfissionalPainel, 0)

	for rows.Next() {
		var p ProfissionalPainel

		if err := rows.Scan(
			&p.ID,
			&p.Nome,
			&p.Especialidade,
			&p.FotoURL,
		); err != nil {
			return nil, err
		}

		profissionais = append(profissionais, p)
	}

	return profissionais, rows.Err()
}

// ListarAgendamentos devolve os agendamentos do dia já agrupados
// (1 item por agendamento, com os serviços dentro).
//
// Cliente:
//   - agendamento feito pelo site/app: auth_user_id preenchido
//     (telefone vem do vínculo, se não houver nome_cliente);
//   - agendamento avulso: nome_cliente e telefone_cliente da própria linha.
func (r *Repository) ListarAgendamentos(
	ctx context.Context,
	salaoID int64,
	data string,
	profissionalID int64,
) ([]AgendamentoPainel, error) {
	query := `
		SELECT
			a.id,
			TIME_FORMAT(a.hora_marcada, '%H:%i'),
			TIME_FORMAT(
				DATE_ADD(a.hora_marcada, INTERVAL a.duracao_minutos MINUTE),
				'%H:%i'
			),
			a.duracao_minutos,
			a.status,
			a.observacao,
			a.profissional_id,
			p.nome,
			a.auth_user_id,
			a.nome_cliente,
			COALESCE(a.telefone_cliente, v.telefone),
			COALESCE(spa.nome_servico_snapshot, ''),
			COALESCE(spa.preco_unitario, 0),
			COALESCE(spa.duracao_minutos, 0),
			COALESCE(spa.quantidade, 1)
		FROM agendamentos a
		INNER JOIN profissionais p
			ON p.id = a.profissional_id
		LEFT JOIN clientes_auth_vinculos v
			ON v.auth_user_id = a.auth_user_id
		LEFT JOIN servico_por_agendamento spa
			ON spa.agendamento_id = a.id
		WHERE a.salao_id = ?
		  AND a.hora_marcada >= ?
		  AND a.hora_marcada < DATE_ADD(?, INTERVAL 1 DAY)`

	args := []any{salaoID, data, data}

	if profissionalID > 0 {
		query += ` AND a.profissional_id = ?`
		args = append(args, profissionalID)
	}

	query += `
		ORDER BY a.hora_marcada ASC, a.id ASC, spa.id ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agendamentos := make([]AgendamentoPainel, 0)
	indice := make(map[int64]int)

	for rows.Next() {
		var (
			id             int64
			hora           string
			horaFim        string
			duracaoMinutos int
			status         string
			observacao     sql.NullString
			profissionalID int64
			profissional   string
			authUserID     sql.NullInt64
			nomeCliente    sql.NullString
			telefone       sql.NullString
			servicoNome    string
			servicoPreco   float64
			servicoDuracao int
			servicoQtd     int
		)

		if err := rows.Scan(
			&id,
			&hora,
			&horaFim,
			&duracaoMinutos,
			&status,
			&observacao,
			&profissionalID,
			&profissional,
			&authUserID,
			&nomeCliente,
			&telefone,
			&servicoNome,
			&servicoPreco,
			&servicoDuracao,
			&servicoQtd,
		); err != nil {
			return nil, err
		}

		pos, ok := indice[id]

		if !ok {
			agendamentos = append(agendamentos, AgendamentoPainel{
				ID:             id,
				Hora:           hora,
				HoraFim:        horaFim,
				DuracaoMinutos: duracaoMinutos,
				Status:         status,
				ProfissionalID: profissionalID,
				Profissional:   profissional,
				Cliente:        montarCliente(authUserID, nomeCliente, telefone),
				Servicos:       []ServicoPainel{},
				Observacao:     strings.TrimSpace(observacao.String),
			})

			pos = len(agendamentos) - 1
			indice[id] = pos
		}

		if servicoNome != "" {
			agendamentos[pos].Servicos = append(
				agendamentos[pos].Servicos,
				ServicoPainel{
					Nome:       servicoNome,
					Quantidade: servicoQtd,
					Preco:      servicoPreco,
					DuracaoMin: servicoDuracao,
				},
			)

			agendamentos[pos].PrecoTotal +=
				servicoPreco * float64(servicoQtd)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return agendamentos, nil
}

func montarCliente(
	authUserID sql.NullInt64,
	nomeCliente sql.NullString,
	telefone sql.NullString,
) ClientePainel {
	nome := strings.TrimSpace(nomeCliente.String)

	cliente := ClientePainel{
		Nome:     nome,
		Telefone: strings.TrimSpace(telefone.String),
		Origem:   "avulso",
	}

	if authUserID.Valid {
		cliente.Origem = "app"

		if cliente.Nome == "" {
			cliente.Nome = "Cliente do app"
		}

		return cliente
	}

	if cliente.Nome == "" {
		cliente.Nome = "Cliente avulso"
	}

	return cliente
}