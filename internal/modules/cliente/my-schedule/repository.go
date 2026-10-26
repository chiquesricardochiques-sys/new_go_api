package myschedule

import (
	"context"
	"database/sql"
	"time"
	"fmt"
)

const layoutDataHora = "2006-01-02 15:04:05"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BuscarTextos(ctx context.Context, salaoID int64) (map[string]string, error) {
	const query = `
		SELECT chave, conteudo
		FROM textos_salao
		WHERE salao_id = ?
		ORDER BY chave ASC`

	rows, err := r.db.QueryContext(ctx, query, salaoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	textos := make(map[string]string)
	for rows.Next() {
		var chave, conteudo string
		if err := rows.Scan(&chave, &conteudo); err != nil {
			return nil, err
		}
		textos[chave] = conteudo
	}
	return textos, rows.Err()
}

// ListarAgendamentos devolve os agendamentos do cliente já agrupados
// (1 item por agendamento, com os serviços dentro).
//
//	meus_agendamentos: começam hoje ou depois e ainda estão ativos
//	                   (pendente/confirmado)
//	historico:         já passaram OU estão concluído/cancelado/não compareceu
func (r *Repository) ListarAgendamentos(
	ctx context.Context,
	salaoID, authUserID int64,
	cenario, status string,
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
			COALESCE(spa.nome_servico_snapshot, ''),
			COALESCE(spa.preco_unitario, 0),
			COALESCE(spa.duracao_minutos, 0),
			COALESCE(spa.quantidade, 1)
		FROM agendamentos a
		INNER JOIN profissionais p
			ON p.id = a.profissional_id
		LEFT JOIN servico_por_agendamento spa
			ON spa.agendamento_id = a.id
		WHERE a.salao_id = ?
		  AND a.auth_user_id = ?`

	args := []any{
		salaoID,
		authUserID,
	}

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
		  )`

		args = append(args, hoje)

		ordem = `
			ORDER BY
				a.hora_marcada DESC,
				a.id DESC,
				spa.id ASC`

	default:
		query += `
		  AND a.hora_marcada >= ?
		  AND a.status IN (
				'pendente',
				'confirmado'
		  )`

		args = append(args, hoje)

		ordem = `
			ORDER BY
				a.hora_marcada ASC,
				a.id ASC,
				spa.id ASC`
	}

	if status != "" {
		query += ` AND a.status = ?`
		args = append(args, status)
	}

	query += ordem

	rows, err := r.db.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agendamentos := make([]Agendamento, 0)

	indice := make(map[int64]int)

	for rows.Next() {

		var (
			id                int64
			inicioStr         string
			duracaoMinutos    int
			st                string
			observacao        sql.NullString
			profissional      string
			servicoNome       string
			servicoPreco      float64
			servicoDuracao    int
			servicoQuantidade int
		)

		if err := rows.Scan(
			&id,
			&inicioStr,
			&duracaoMinutos,
			&st,
			&observacao,
			&profissional,
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

			/*
				Importante:

				hora_marcada = início real do agendamento
				duracao_minutos = duração real

				O fim é calculado aqui.
			*/
			fim := inicio.Add(
				time.Duration(duracaoMinutos) * time.Minute,
			)

			ag := Agendamento{
				ID:           id,
				Data:         inicio.Format("2006-01-02"),
				Hora:         inicio.Format("15:04"),
				HoraFim:      fim.Format("15:04"),
				Status:       st,
				Profissional: profissional,
				Servicos:     []ServicoAgendado{},
				Inicio:       inicio,
			}

			if observacao.Valid {
				obs := observacao.String
				ag.Observacoes = &obs
			}

			agendamentos = append(
				agendamentos,
				ag,
			)

			pos = len(agendamentos) - 1

			indice[id] = pos
		}

		/*
			Os serviços continuam vindo do snapshot
			do momento em que o agendamento foi criado.
		*/
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

// CancelarAgendamento muda o status para 'cancelado' (não apaga a linha).
// Só cancela se o agendamento é do cliente, está pendente/confirmado
// e ainda não começou. Tudo dentro de uma transação.
func (r *Repository) CancelarAgendamento(
	ctx context.Context,
	salaoID, authUserID, agendamentoID int64,
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
		  AND salao_id = ?
		  AND auth_user_id = ?
		LIMIT 1
		FOR UPDATE`

	var (
		status    string
		inicioStr string
	)

	err = tx.QueryRowContext(
		ctx,
		selectQuery,
		agendamentoID,
		salaoID,
		authUserID,
	).Scan(
		&status,
		&inicioStr,
	)

	if err != nil {
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

	if status != "pendente" &&
		status != "confirmado" {
		return fmt.Errorf("agendamento não pode ser cancelado")
	}

	if !inicio.After(agora) {
		return fmt.Errorf("agendamento não pode ser cancelado")
	}

	const updateQuery = `
		UPDATE agendamentos
		SET
			status = 'cancelado',
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		  AND salao_id = ?
		  AND auth_user_id = ?`

	_, err = tx.ExecContext(
		ctx,
		updateQuery,
		agendamentoID,
		salaoID,
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