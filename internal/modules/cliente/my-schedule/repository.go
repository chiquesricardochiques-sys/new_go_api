package myschedule

import (
	"context"
	"database/sql"
	"time"
)

const layoutDataHora = "2006-01-02 15:04:05"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ============================================================
// LISTAR AGENDAMENTOS GLOBAIS
// ============================================================

func (r *Repository) ListarAgendamentos(
	ctx context.Context,
	authUserID int64,
	cenario string,
	status string,
	inicioHoje time.Time,
) ([]Agendamento, error) {

	query := `
		SELECT
			a.id,
			DATE_FORMAT(a.hora_marcada, '%Y-%m-%d %H:%i:%s'),
			a.duracao_minutos,
			a.status,
			a.observacao,
			p.nome,
			COALESCE(s.nome, 'Salão'),
			COALESCE(spa.nome_servico_snapshot, ''),
			COALESCE(spa.preco_unitario, 0),
			COALESCE(spa.duracao_minutos, 0),
			COALESCE(spa.quantidade, 1)
		FROM agendamentos a

		INNER JOIN profissionais p
			ON p.id = a.profissional_id

		LEFT JOIN saloes s
			ON s.id = a.salao_id

		LEFT JOIN servico_por_agendamento spa
			ON spa.agendamento_id = a.id

		WHERE a.auth_user_id = ?
	`

	args := []any{authUserID}

	hoje := inicioHoje.Format(layoutDataHora)

	var ordem string

	switch cenario {

	case CenarioHistorico:

		query += `
			AND (
				a.hora_marcada < ?
				OR a.status IN (
					'concluido',
					'cancelado',
					'nao_compareceu'
				)
			)
		`

		args = append(args, hoje)

		ordem = `
			ORDER BY
				a.hora_marcada DESC,
				a.id DESC,
				spa.id ASC
		`

	default:

		query += `
			AND a.hora_marcada >= ?
			AND a.status IN (
				'pendente',
				'confirmado'
			)
		`

		args = append(args, hoje)

		ordem = `
			ORDER BY
				a.hora_marcada ASC,
				a.id ASC,
				spa.id ASC
		`
	}

	if status != "" {
		query += ` AND a.status = ?`
		args = append(args, status)
	}

	query += ordem

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	agendamentos := make([]Agendamento, 0)
	indice := make(map[int64]int)

	for rows.Next() {

		var (
			id                 int64
			inicioStr           string
			duracaoMinutos     int
			st                  string
			observacao          sql.NullString
			profissional        string
			nomeSalao          string
			servicoNome         string
			servicoPreco       float64
			servicoDuracao     int
			servicoQuantidade  int
		)

		if err := rows.Scan(
			&id,
			&inicioStr,
			&duracaoMinutos,
			&st,
			&observacao,
			&profissional,
			&nomeSalao,
			&servicoNome,
			&servicoPreco,
			&servicoDuracao,
			&servicoQuantidade,
		); err != nil {
			return nil, err
		}

		pos, ok := indice[id]

		if !ok {

			inicio, err := time.ParseInLocation(
				layoutDataHora,
				inicioStr,
				localBR,
			)

			if err != nil {
				return nil, err
			}

			fim := inicio.Add(
				time.Duration(duracaoMinutos) * time.Minute,
			)

			ag := Agendamento{
				ID:           id,
				Salao:        nomeSalao,
				Data:          inicio.Format("2006-01-02"),
				Hora:          inicio.Format("15:04"),
				HoraFim:       fim.Format("15:04"),
				Status:         st,
				Profissional:  profissional,
				Servicos:      []ServicoAgendado{},
				Inicio:        inicio,
			}

			if observacao.Valid {
				obs := observacao.String
				ag.Observacoes = &obs
			}

			agendamentos = append(agendamentos, ag)

			pos = len(agendamentos) - 1
			indice[id] = pos
		}

		if servicoNome != "" {

			agendamentos[pos].Servicos = append(
				agendamentos[pos].Servicos,
				ServicoAgendado{
					Nome:       servicoNome,
					Preco:      servicoPreco,
					DuracaoMin: servicoDuracao,
					Quantidade: servicoQuantidade,
				},
			)

			agendamentos[pos].DuracaoTotal +=
				servicoDuracao * servicoQuantidade

			agendamentos[pos].PrecoTotal +=
				servicoPreco * float64(servicoQuantidade)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return agendamentos, nil
}

// ============================================================
// CANCELAR AGENDAMENTO GLOBAL
// ============================================================

func (r *Repository) CancelarAgendamento(
	ctx context.Context,
	authUserID int64,
	agendamentoID int64,
	agora time.Time,
) error {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	const selectQuery = `
		SELECT
			status,
			DATE_FORMAT(
				hora_marcada,
				'%Y-%m-%d %H:%i:%s'
			)
		FROM agendamentos
		WHERE id = ?
		  AND auth_user_id = ?
		LIMIT 1
		FOR UPDATE
	`

	var (
		status    string
		inicioStr string
	)

	err = tx.QueryRowContext(
		ctx,
		selectQuery,
		agendamentoID,
		authUserID,
	).Scan(&status, &inicioStr)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAgendamentoNaoEncontrado
		}

		return err
	}

	inicio, err := time.ParseInLocation(
		layoutDataHora,
		inicioStr,
		localBR,
	)

	if err != nil {
		return err
	}

	if status == "cancelado" {
		return ErrJaCancelado
	}

	if status != "pendente" && status != "confirmado" {
		return ErrNaoPodeCancelar
	}

	if !inicio.After(agora) {
		return ErrNaoPodeCancelar
	}

	const updateQuery = `
		UPDATE agendamentos
		SET
			status = 'cancelado',
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		  AND auth_user_id = ?
	`

	_, err = tx.ExecContext(
		ctx,
		updateQuery,
		agendamentoID,
		authUserID,
	)

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}