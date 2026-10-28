package profissional

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

// ============================================================================
// PROFISSIONAL
// ============================================================================

const colunasProfissional = `
	p.id, p.salao_id, p.auth_user_id, p.nome,
	COALESCE(p.especialidade, ''), COALESCE(p.foto_url, ''), p.status,
	DATE_FORMAT(p.created_at, '%Y-%m-%d %H:%i:%s'),
	p.tipo_remuneracao, p.percentual_comissao, p.salario_fixo,
	p.frequencia_pagamento, p.usa_horarios`

type scanner interface {
	Scan(dest ...any) error
}

func scanProfissional(row scanner) (*Profissional, error) {
	var p Profissional
	var auth sql.NullInt64

	err := row.Scan(
		&p.ID, &p.IDSalao, &auth, &p.Nome,
		&p.Especialidade, &p.Img, &p.Status,
		&p.CriadoEm,
		&p.TipoRemuneracao, &p.PercentualComissao, &p.SalarioFixo,
		&p.FrequenciaPagamento, &p.UsaHorarios,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNaoEncontrado
	}
	if err != nil {
		return nil, err
	}

	if auth.Valid {
		v := auth.Int64
		p.AuthUserID = &v
	}
	if p.Status == "ativo" {
		p.Ativo = 1
	}

	return &p, nil
}

func (r *Repository) BuscarPorID(ctx context.Context, idSalao, profissionalID int64) (*Profissional, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+colunasProfissional+`
		 FROM profissionais p
		 WHERE p.id = ? AND p.salao_id = ? AND p.deleted_at IS NULL
		 LIMIT 1`,
		profissionalID, idSalao,
	)
	return scanProfissional(row)
}

func (r *Repository) BuscarPorAuthUser(ctx context.Context, idSalao, authUserID int64) (*Profissional, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+colunasProfissional+`
		 FROM profissionais p
		 WHERE p.auth_user_id = ? AND p.salao_id = ? AND p.deleted_at IS NULL
		 LIMIT 1`,
		authUserID, idSalao,
	)
	return scanProfissional(row)
}

// ============================================================================
// AGENDAMENTOS (agenda / histórico / horários do dia)
// ============================================================================

// ListarAgendamentos devolve os agendamentos já agrupados (1 item por agendamento,
// com a lista de serviços). Sempre filtra por salao_id + profissional_id.
func (r *Repository) ListarAgendamentos(ctx context.Context, f FiltroAgendamentos) ([]AgendamentoItem, error) {
	query := `
		SELECT
			a.id,
			DATE_FORMAT(a.inicio, '%Y-%m-%d'),
			DATE_FORMAT(a.inicio, '%H:%i'),
			DATE_FORMAT(a.fim, '%H:%i'),
			a.status,
			COALESCE(a.observacao, ''),
			a.auth_user_id,
			COALESCE(a.percentual_comissao, p.percentual_comissao),
			spa.nome_servico_snapshot,
			spa.duracao_minutos,
			spa.quantidade,
			spa.preco_unitario
		FROM agendamentos a
		JOIN profissionais p ON p.id = a.profissional_id
		LEFT JOIN servico_por_agendamento spa ON spa.agendamento_id = a.id
		WHERE a.salao_id = ? AND a.profissional_id = ?`

	args := []any{f.IDSalao, f.ProfissionalID}

	if f.Inicio != "" {
		query += ` AND a.inicio >= ?`
		args = append(args, f.Inicio)
	}
	if f.Fim != "" {
		query += ` AND a.inicio < ?`
		args = append(args, f.Fim)
	}
	if f.Status != "" {
		query += ` AND a.status = ?`
		args = append(args, f.Status)
	}

	if f.Desc {
		query += ` ORDER BY a.inicio DESC, a.id DESC, spa.id ASC`
	} else {
		query += ` ORDER BY a.inicio ASC, a.id ASC, spa.id ASC`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itens := make(map[int64]*AgendamentoItem)
	ordem := make([]int64, 0)

	for rows.Next() {
		var (
			id, clienteID                int64
			data, hora, horaFim, status  string
			observacao                   string
			comissao                     float64
			nomeServico                  sql.NullString
			duracao, quantidade          sql.NullInt64
			preco                        sql.NullFloat64
		)

		if err := rows.Scan(
			&id, &data, &hora, &horaFim, &status, &observacao, &clienteID,
			&comissao, &nomeServico, &duracao, &quantidade, &preco,
		); err != nil {
			return nil, err
		}

		item, existe := itens[id]
		if !existe {
			item = &AgendamentoItem{
				ID:                 id,
				Data:               data,
				Hora:               hora,
				HoraFim:            horaFim,
				Status:             status,
				Observacoes:        observacao,
				ClienteID:          clienteID,
				Cliente:            "N/A",
				Servicos:           []string{},
				PercentualComissao: comissao,
			}
			itens[id] = item
			ordem = append(ordem, id)
		}

		if nomeServico.Valid {
			qtd := int64(1)
			if quantidade.Valid && quantidade.Int64 > 0 {
				qtd = quantidade.Int64
			}
			item.Servicos = append(item.Servicos, nomeServico.String)
			item.DuracaoTotal += int(duracao.Int64 * qtd)
			item.PrecoTotal += preco.Float64 * float64(qtd)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	resultado := make([]AgendamentoItem, 0, len(ordem))
	for _, id := range ordem {
		resultado = append(resultado, *itens[id])
	}

	return resultado, nil
}

func (r *Repository) ContarPeriodo(ctx context.Context, idSalao, profissionalID int64, inicio, fim string) (int, error) {
	var total int

	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM agendamentos
		WHERE salao_id = ?
		  AND profissional_id = ?
		  AND status <> 'cancelado'
		  AND inicio >= ?
		  AND inicio < ?`,
		idSalao, profissionalID, inicio, fim,
	).Scan(&total)

	return total, err
}

func (r *Repository) TotaisHistorico(ctx context.Context, idSalao, profissionalID int64) (concluidos, cancelados int, err error) {
	err = r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(status = 'concluido'), 0),
			COALESCE(SUM(status = 'cancelado'), 0)
		FROM agendamentos
		WHERE salao_id = ? AND profissional_id = ?`,
		idSalao, profissionalID,
	).Scan(&concluidos, &cancelados)

	return
}

// ============================================================================
// ALTERAR STATUS
// ============================================================================

func (r *Repository) StatusDoAgendamento(ctx context.Context, idSalao, profissionalID, agendamentoID int64) (string, error) {
	var status string

	err := r.db.QueryRowContext(ctx, `
		SELECT status
		FROM agendamentos
		WHERE id = ? AND salao_id = ? AND profissional_id = ?
		LIMIT 1`,
		agendamentoID, idSalao, profissionalID,
	).Scan(&status)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNaoEncontrado
	}

	return status, err
}

// AtualizarStatus muda o status e, ao concluir, congela (snapshot) o percentual de
// comissão do profissional no agendamento — o equivalente ao que o Node fazia
// em "andamento". Requer a coluna agendamentos.percentual_comissao.
func (r *Repository) AtualizarStatus(ctx context.Context, idSalao, profissionalID, agendamentoID int64, novoStatus string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE agendamentos a
		JOIN profissionais p ON p.id = a.profissional_id
		SET a.status = ?,
		    a.percentual_comissao = IF(
		        ? = 'concluido',
		        COALESCE(a.percentual_comissao, p.percentual_comissao),
		        a.percentual_comissao
		    )
		WHERE a.id = ? AND a.salao_id = ? AND a.profissional_id = ?`,
		novoStatus, novoStatus, agendamentoID, idSalao, profissionalID,
	)
	return err
}

// ============================================================================
// GRADE / BLOQUEIOS
// ============================================================================

func (r *Repository) BuscarGrade(ctx context.Context, idSalao, profissionalID int64, diaSemana string) (*Grade, error) {
	var (
		g    Grade
		abre int
	)

	err := r.db.QueryRowContext(ctx, `
		SELECT
			abre,
			COALESCE(TIME_FORMAT(abertura, '%H:%i'), ''),
			COALESCE(TIME_FORMAT(pausa_inicio, '%H:%i'), ''),
			COALESCE(TIME_FORMAT(pausa_fim, '%H:%i'), ''),
			COALESCE(TIME_FORMAT(fechamento, '%H:%i'), '')
		FROM profissional_grade_semanal
		WHERE salao_id = ? AND profissional_id = ? AND dia_semana = ?
		LIMIT 1`,
		idSalao, profissionalID, diaSemana,
	).Scan(&abre, &g.Abertura, &g.PausaInicio, &g.PausaFim, &g.Fechamento)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	g.Abre = abre == 1
	return &g, nil
}

func (r *Repository) DiaBloqueado(ctx context.Context, profissionalID int64, data string) (bool, error) {
	var bloqueado bool

	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM profissional_dias_bloqueados
			WHERE profissional_id = ? AND data = ?
		)`,
		profissionalID, data,
	).Scan(&bloqueado)

	return bloqueado, err
}

func (r *Repository) BloqueiosDoDia(ctx context.Context, profissionalID int64, inicioDia, fimDia string) ([]BloqueioHorario, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			DATE_FORMAT(inicio, '%Y-%m-%d %H:%i'),
			DATE_FORMAT(fim, '%Y-%m-%d %H:%i')
		FROM profissional_horarios_bloqueados
		WHERE profissional_id = ? AND inicio < ? AND fim > ?`,
		profissionalID, fimDia, inicioDia,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bloqueios []BloqueioHorario
	for rows.Next() {
		var b BloqueioHorario
		if err := rows.Scan(&b.Inicio, &b.Fim); err != nil {
			return nil, err
		}
		bloqueios = append(bloqueios, b)
	}

	return bloqueios, rows.Err()
}

// ============================================================================
// FATURAMENTO
// ============================================================================

// LinhasFaturamento devolve 1 linha por agendamento do período (exceto cancelados
// e não comparecimentos), com o preço total e o percentual de comissão aplicável
// (snapshot do agendamento, ou o atual do profissional se ainda não houver).
func (r *Repository) LinhasFaturamento(ctx context.Context, idSalao, profissionalID int64, inicio, fim string) ([]LinhaFaturamento, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			a.id,
			COALESCE((
				SELECT SUM(spa.preco_unitario * spa.quantidade)
				FROM servico_por_agendamento spa
				WHERE spa.agendamento_id = a.id
			), 0),
			COALESCE(a.percentual_comissao, p.percentual_comissao)
		FROM agendamentos a
		JOIN profissionais p ON p.id = a.profissional_id
		WHERE a.salao_id = ?
		  AND a.profissional_id = ?
		  AND a.status NOT IN ('cancelado', 'nao_compareceu')
		  AND a.inicio >= ?
		  AND a.inicio < ?`,
		idSalao, profissionalID, inicio, fim,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var linhas []LinhaFaturamento
	for rows.Next() {
		var l LinhaFaturamento
		if err := rows.Scan(&l.ID, &l.PrecoTotal, &l.PercentualComissao); err != nil {
			return nil, err
		}
		linhas = append(linhas, l)
	}

	return linhas, rows.Err()
}