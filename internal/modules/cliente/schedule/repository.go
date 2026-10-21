package schedule

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

const layoutDataHora = "2006-01-02 15:04:05"

// Datas/horas vão e voltam do banco como texto "YYYY-MM-DD HH:MM:SS"
// (horário de parede). Assim não depende de parseTime/loc do DSN.
func dataHoraSQL(t time.Time) string {
	return t.Format(layoutDataHora)
}

func parseDataHora(s string) (time.Time, error) {
	return time.ParseInLocation(layoutDataHora, s, localBR)
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) VerificarSalao(ctx context.Context, salaoID int64) error {
	const query = `SELECT id FROM saloes WHERE id = ? AND status = 'ativo' LIMIT 1`

	var id int64
	err := r.db.QueryRowContext(ctx, query, salaoID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSalaoNaoEncontrado
	}
	return err
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

func (r *Repository) BuscarServicos(ctx context.Context, salaoID int64) ([]Servico, error) {
	const query = `
		SELECT s.id, s.nome, s.descricao, ss.preco, ss.duracao_minutos
		FROM salao_servicos ss
		INNER JOIN servicos s ON s.id = ss.servico_id
		WHERE ss.salao_id = ?
		  AND ss.status = 'ativo'
		  AND s.status = 'ativo'
		ORDER BY s.nome ASC`

	rows, err := r.db.QueryContext(ctx, query, salaoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servicos := make([]Servico, 0)
	for rows.Next() {
		var s Servico
		if err := rows.Scan(&s.ID, &s.Nome, &s.Descricao, &s.Preco, &s.DuracaoMinutos); err != nil {
			return nil, err
		}
		servicos = append(servicos, s)
	}
	return servicos, rows.Err()
}

func (r *Repository) BuscarProfissionais(ctx context.Context, salaoID int64) ([]Profissional, error) {
	const query = `
		SELECT id, nome, especialidade, foto_url, usa_horarios
		FROM profissionais
		WHERE salao_id = ?
		  AND status = 'ativo'
		  AND deleted_at IS NULL
		ORDER BY nome ASC`

	rows, err := r.db.QueryContext(ctx, query, salaoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profissionais := make([]Profissional, 0)
	for rows.Next() {
		var p Profissional
		var usaHorarios int
		if err := rows.Scan(&p.ID, &p.Nome, &p.Especialidade, &p.FotoURL, &usaHorarios); err != nil {
			return nil, err
		}
		p.UsaHorarios = usaHorarios == 1
		profissionais = append(profissionais, p)
	}
	return profissionais, rows.Err()
}

func (r *Repository) BuscarProfissional(ctx context.Context, salaoID, profissionalID int64) (*Profissional, error) {
	const query = `
		SELECT id, nome, especialidade, foto_url, usa_horarios
		FROM profissionais
		WHERE id = ?
		  AND salao_id = ?
		  AND status = 'ativo'
		  AND deleted_at IS NULL
		LIMIT 1`

	var p Profissional
	var usaHorarios int
	err := r.db.QueryRowContext(ctx, query, profissionalID, salaoID).
		Scan(&p.ID, &p.Nome, &p.Especialidade, &p.FotoURL, &usaHorarios)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProfissionalNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	p.UsaHorarios = usaHorarios == 1
	return &p, nil
}

func (r *Repository) BuscarGrade(ctx context.Context, salaoID, profissionalID int64, diaSemana string) (*Grade, error) {
	const query = `
		SELECT abre, abertura, pausa_inicio, pausa_fim, fechamento
		FROM profissional_grade_semanal
		WHERE salao_id = ?
		  AND profissional_id = ?
		  AND dia_semana = ?
		LIMIT 1`

	var (
		abre                  int
		abertura, fechamento  sql.NullString
		pausaInicio, pausaFim sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, salaoID, profissionalID, diaSemana).
		Scan(&abre, &abertura, &pausaInicio, &pausaFim, &fechamento)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrGradeNaoEncontrada
	}
	if err != nil {
		return nil, err
	}

	g := &Grade{Abre: abre == 1}

	// dia fechado: abertura/fechamento podem ser NULL
	if !g.Abre || !abertura.Valid || !fechamento.Valid {
		g.Abre = false
		return g, nil
	}

	g.Abertura = abertura.String
	g.Fechamento = fechamento.String
	if pausaInicio.Valid && pausaFim.Valid {
		pi, pf := pausaInicio.String, pausaFim.String
		g.PausaInicio, g.PausaFim = &pi, &pf
	}
	return g, nil
}

func (r *Repository) VerificarDiaBloqueado(ctx context.Context, profissionalID int64, data string) (bool, error) {
	const query = `
		SELECT id
		FROM profissional_dias_bloqueados
		WHERE profissional_id = ? AND data = ?
		LIMIT 1`

	var id int64
	err := r.db.QueryRowContext(ctx, query, profissionalID, data).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) VerificarHorarioBloqueado(ctx context.Context, profissionalID int64, inicio, fim time.Time) (bool, error) {
	const query = `
		SELECT id
		FROM profissional_horarios_bloqueados
		WHERE profissional_id = ?
		  AND inicio < ?
		  AND fim > ?
		LIMIT 1`

	var id int64
	err := r.db.QueryRowContext(ctx, query, profissionalID, dataHoraSQL(fim), dataHoraSQL(inicio)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// Bloqueios de horário que encostam no dia.
func (r *Repository) BuscarBloqueiosDoDia(ctx context.Context, profissionalID int64, inicioDia, fimDia time.Time) ([]BloqueioHorario, error) {
	const query = `
		SELECT
			DATE_FORMAT(inicio, '%Y-%m-%d %H:%i:%s'),
			DATE_FORMAT(fim,    '%Y-%m-%d %H:%i:%s')
		FROM profissional_horarios_bloqueados
		WHERE profissional_id = ?
		  AND inicio < ?
		  AND fim > ?
		ORDER BY inicio ASC`

	rows, err := r.db.QueryContext(ctx, query, profissionalID, dataHoraSQL(fimDia), dataHoraSQL(inicioDia))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bloqueios := make([]BloqueioHorario, 0)
	for rows.Next() {
		var ini, fim string
		if err := rows.Scan(&ini, &fim); err != nil {
			return nil, err
		}
		ti, err := parseDataHora(ini)
		if err != nil {
			return nil, err
		}
		tf, err := parseDataHora(fim)
		if err != nil {
			return nil, err
		}
		bloqueios = append(bloqueios, BloqueioHorario{Inicio: ti, Fim: tf})
	}
	return bloqueios, rows.Err()
}

// Agendamentos ativos (pendente/confirmado) do profissional no dia,
// já com os nomes dos serviços de cada um.
func (r *Repository) BuscarAgendamentosDoDia(ctx context.Context, salaoID, profissionalID int64, inicioDia, fimDia time.Time) ([]AgendamentoExistente, error) {
	const query = `
		SELECT
			a.id,
			DATE_FORMAT(a.inicio, '%Y-%m-%d %H:%i:%s'),
			DATE_FORMAT(a.fim,    '%Y-%m-%d %H:%i:%s'),
			COALESCE(spa.nome_servico_snapshot, '')
		FROM agendamentos a
		LEFT JOIN servico_por_agendamento spa ON spa.agendamento_id = a.id
		WHERE a.salao_id = ?
		  AND a.profissional_id = ?
		  AND a.status IN ('pendente', 'confirmado')
		  AND a.inicio < ?
		  AND a.fim > ?
		ORDER BY a.inicio ASC, a.id ASC, spa.id ASC`

	rows, err := r.db.QueryContext(ctx, query,
		salaoID, profissionalID, dataHoraSQL(fimDia), dataHoraSQL(inicioDia))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agendamentos := make([]AgendamentoExistente, 0)
	indice := make(map[int64]int)

	for rows.Next() {
		var id int64
		var ini, fim, servico string
		if err := rows.Scan(&id, &ini, &fim, &servico); err != nil {
			return nil, err
		}

		pos, ok := indice[id]
		if !ok {
			ti, err := parseDataHora(ini)
			if err != nil {
				return nil, err
			}
			tf, err := parseDataHora(fim)
			if err != nil {
				return nil, err
			}
			agendamentos = append(agendamentos, AgendamentoExistente{ID: id, Inicio: ti, Fim: tf})
			pos = len(agendamentos) - 1
			indice[id] = pos
		}
		if servico != "" {
			agendamentos[pos].Servicos = append(agendamentos[pos].Servicos, servico)
		}
	}
	return agendamentos, rows.Err()
}

// O cliente é identificado pelo auth_user_id (vem da sessão no front).
func (r *Repository) VerificarCliente(ctx context.Context, authUserID int64) error {
	const query = `SELECT id FROM clientes_auth_vinculos WHERE auth_user_id = ? LIMIT 1`

	var id int64
	err := r.db.QueryRowContext(ctx, query, authUserID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrClienteNaoEncontrado
	}
	return err
}

// Serviços do salão que esse profissional realmente atende.
// Retorna erro se algum ID não for válido.
func (r *Repository) BuscarServicosSelecionados(ctx context.Context, salaoID, profissionalID int64, servicoIDs []int64) ([]Servico, int, error) {
	// remove IDs repetidos
	vistos := make(map[int64]bool, len(servicoIDs))
	ids := make([]int64, 0, len(servicoIDs))
	for _, id := range servicoIDs {
		if id <= 0 {
			return nil, 0, ErrServicoNaoEncontrado
		}
		if !vistos[id] {
			vistos[id] = true
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil, 0, ErrServicoNaoEncontrado
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")

	query := `
		SELECT s.id, s.nome, s.descricao, ss.preco, ss.duracao_minutos
		FROM salao_servicos ss
		INNER JOIN servicos s ON s.id = ss.servico_id
		INNER JOIN servico_profissional sp
			ON sp.servico_id = ss.servico_id
		   AND sp.profissional_id = ?
		   AND sp.salao_id = ?
		   AND sp.status = 'ativo'
		WHERE ss.salao_id = ?
		  AND ss.status = 'ativo'
		  AND s.status = 'ativo'
		  AND ss.servico_id IN (` + placeholders + `)
		ORDER BY s.nome ASC`

	args := make([]any, 0, len(ids)+3)
	args = append(args, profissionalID, salaoID, salaoID)
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	servicos := make([]Servico, 0, len(ids))
	duracaoTotal := 0
	for rows.Next() {
		var s Servico
		if err := rows.Scan(&s.ID, &s.Nome, &s.Descricao, &s.Preco, &s.DuracaoMinutos); err != nil {
			return nil, 0, err
		}
		servicos = append(servicos, s)
		duracaoTotal += s.DuracaoMinutos
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(servicos) != len(ids) {
		return nil, 0, ErrServicoNaoEncontrado
	}
	return servicos, duracaoTotal, nil
}

// CriarAgendamento faz a checagem final de "já tem gente nesse horário"
// DENTRO da transação. Trava a linha do profissional pra duas requisições
// simultâneas não passarem juntas pela checagem.
func (r *Repository) CriarAgendamento(
	ctx context.Context,
	req AgendarRequest,
	servicos []Servico,
	inicio, fim time.Time,
) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	// 1. trava o profissional (serializa agendamentos dele)
	const lockQuery = `
		SELECT id
		FROM profissionais
		WHERE id = ?
		  AND salao_id = ?
		  AND status = 'ativo'
		  AND deleted_at IS NULL
		LIMIT 1
		FOR UPDATE`

	var profissionalID int64
	err = tx.QueryRowContext(ctx, lockQuery, req.ProfissionalID, req.SalaoID).Scan(&profissionalID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrProfissionalNaoEncontrado
	}
	if err != nil {
		return 0, err
	}

	// 2. já existe agendamento ativo sobrepondo esse intervalo?
	const conflitoQuery = `
		SELECT id
		FROM agendamentos
		WHERE salao_id = ?
		  AND profissional_id = ?
		  AND status IN ('pendente', 'confirmado')
		  AND inicio < ?
		  AND fim > ?
		LIMIT 1`

	var conflitoID int64
	err = tx.QueryRowContext(ctx, conflitoQuery,
		req.SalaoID, req.ProfissionalID, dataHoraSQL(fim), dataHoraSQL(inicio)).Scan(&conflitoID)
	if err == nil {
		return 0, ErrHorarioIndisponivel
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	// 3. cria o agendamento
	const insertQuery = `
		INSERT INTO agendamentos
			(salao_id, profissional_id, auth_user_id, inicio, fim, status, observacao)
		VALUES (?, ?, ?, ?, ?, 'pendente', ?)`

	result, err := tx.ExecContext(ctx, insertQuery,
		req.SalaoID, req.ProfissionalID, req.AuthUserID,
		dataHoraSQL(inicio), dataHoraSQL(fim), req.Observacao)
	if err != nil {
		return 0, err
	}

	agendamentoID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// 4. snapshot dos serviços
	const snapshotQuery = `
		INSERT INTO servico_por_agendamento
			(agendamento_id, servico_id, nome_servico_snapshot, preco_unitario, duracao_minutos, quantidade)
		VALUES (?, ?, ?, ?, ?, 1)`

	for _, s := range servicos {
		if _, err := tx.ExecContext(ctx, snapshotQuery,
			agendamentoID, s.ID, s.Nome, s.Preco, s.DuracaoMinutos); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return agendamentoID, nil
}