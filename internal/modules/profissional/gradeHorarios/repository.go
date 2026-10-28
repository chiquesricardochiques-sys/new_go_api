package gradehorarios

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

// Profissional do salão (garante que o id pertence ao id_salao).
func (r *Repository) BuscarProfissional(ctx context.Context, idSalao, profissionalID int64) (*Profissional, error) {
	var p Profissional

	err := r.db.QueryRowContext(ctx, `
		SELECT id, usa_horarios
		FROM profissionais
		WHERE id = ? AND salao_id = ? AND deleted_at IS NULL
		LIMIT 1`,
		profissionalID, idSalao,
	).Scan(&p.ID, &p.UsaHorarios)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNaoEncontrado
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

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

	g.Abre = abre == 1 && g.Abertura != "" && g.Fechamento != ""
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

// Agendamentos ativos (não cancelados / não faltou) que tocam o intervalo,
// já com os nomes dos serviços. Usa hora_marcada + duracao_minutos.
func (r *Repository) AgendamentosNoIntervalo(ctx context.Context, idSalao, profissionalID int64, inicio, fim string) ([]AgendamentoDia, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			a.id,
			DATE_FORMAT(a.hora_marcada, '%Y-%m-%d %H:%i:%s'),
			a.duracao_minutos,
			COALESCE(spa.nome_servico_snapshot, '')
		FROM agendamentos a
		LEFT JOIN servico_por_agendamento spa ON spa.agendamento_id = a.id
		WHERE a.salao_id = ?
		  AND a.profissional_id = ?
		  AND a.status NOT IN ('cancelado', 'nao_compareceu')
		  AND a.hora_marcada < ?
		  AND DATE_ADD(a.hora_marcada, INTERVAL a.duracao_minutos MINUTE) > ?
		ORDER BY a.hora_marcada ASC, a.id ASC, spa.id ASC`,
		idSalao, profissionalID, fim, inicio,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itens := make([]AgendamentoDia, 0)
	indice := make(map[int64]int)

	for rows.Next() {
		var (
			id      int64
			ini     string
			duracao int
			servico string
		)

		if err := rows.Scan(&id, &ini, &duracao, &servico); err != nil {
			return nil, err
		}

		pos, ok := indice[id]
		if !ok {
			itens = append(itens, AgendamentoDia{ID: id, Inicio: ini, Duracao: duracao})
			pos = len(itens) - 1
			indice[id] = pos
		}
		if servico != "" {
			itens[pos].Servicos = append(itens[pos].Servicos, servico)
		}
	}

	return itens, rows.Err()
}

func (r *Repository) BloqueiosNoIntervalo(ctx context.Context, profissionalID int64, inicio, fim string) ([]BloqueioDia, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			DATE_FORMAT(inicio, '%Y-%m-%d %H:%i:%s'),
			DATE_FORMAT(fim,    '%Y-%m-%d %H:%i:%s'),
			motivo
		FROM profissional_horarios_bloqueados
		WHERE profissional_id = ?
		  AND inicio < ?
		  AND fim > ?
		ORDER BY inicio ASC, id ASC`,
		profissionalID, fim, inicio,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bloqueios := make([]BloqueioDia, 0)
	for rows.Next() {
		var (
			b      BloqueioDia
			motivo sql.NullString
		)
		if err := rows.Scan(&b.ID, &b.Inicio, &b.Fim, &motivo); err != nil {
			return nil, err
		}
		if motivo.Valid && motivo.String != "" {
			m := motivo.String
			b.Motivo = &m
		}
		bloqueios = append(bloqueios, b)
	}

	return bloqueios, rows.Err()
}

// CriarBloqueio faz a checagem final DENTRO da transação:
//
//  1. trava a linha do profissional (a mesma que o /agendar do cliente trava,
//     então bloquear e agendar não acontecem ao mesmo tempo);
//  2. se existe agendamento ativo tocando o intervalo -> ErrTemAgendamento;
//  3. se já existe bloqueio tocando o intervalo        -> ErrJaBloqueado;
//  4. grava em profissional_horarios_bloqueados.
func (r *Repository) CriarBloqueio(
	ctx context.Context,
	idSalao, profissionalID int64,
	inicio, fim, motivo string,
) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var id int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM profissionais
		WHERE id = ? AND salao_id = ? AND deleted_at IS NULL
		LIMIT 1
		FOR UPDATE`,
		profissionalID, idSalao,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNaoEncontrado
	}
	if err != nil {
		return 0, err
	}

	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM agendamentos
		WHERE salao_id = ?
		  AND profissional_id = ?
		  AND status NOT IN ('cancelado', 'nao_compareceu')
		  AND hora_marcada < ?
		  AND DATE_ADD(hora_marcada, INTERVAL duracao_minutos MINUTE) > ?
		LIMIT 1`,
		idSalao, profissionalID, fim, inicio,
	).Scan(&id)
	if err == nil {
		return 0, ErrTemAgendamento
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM profissional_horarios_bloqueados
		WHERE profissional_id = ?
		  AND inicio < ?
		  AND fim > ?
		LIMIT 1`,
		profissionalID, fim, inicio,
	).Scan(&id)
	if err == nil {
		return 0, ErrJaBloqueado
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	var motivoSQL sql.NullString
	if motivo != "" {
		motivoSQL = sql.NullString{String: motivo, Valid: true}
	}

	resultado, err := tx.ExecContext(ctx, `
		INSERT INTO profissional_horarios_bloqueados (profissional_id, inicio, fim, motivo)
		VALUES (?, ?, ?, ?)`,
		profissionalID, inicio, fim, motivoSQL,
	)
	if err != nil {
		return 0, err
	}

	bloqueioID, err := resultado.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return bloqueioID, nil
}

// RemoverBloqueio apaga o bloqueio, desde que seja do profissional e do salão.
func (r *Repository) RemoverBloqueio(ctx context.Context, idSalao, profissionalID, bloqueioID int64) error {
	resultado, err := r.db.ExecContext(ctx, `
		DELETE b
		FROM profissional_horarios_bloqueados b
		JOIN profissionais p ON p.id = b.profissional_id
		WHERE b.id = ?
		  AND b.profissional_id = ?
		  AND p.salao_id = ?`,
		bloqueioID, profissionalID, idSalao,
	)
	if err != nil {
		return err
	}

	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas == 0 {
		return ErrNaoEncontrado
	}

	return nil
}