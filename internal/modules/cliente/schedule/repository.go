package schedule

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var (
	ErrSalaoNaoEncontrado        = errors.New("salão não encontrado")
	ErrProfissionalNaoEncontrado = errors.New("profissional não encontrado")
	ErrClienteNaoEncontrado      = errors.New("cliente não encontrado")
	ErrServicoNaoEncontrado      = errors.New("serviço não encontrado")
	ErrGradeNaoEncontrada        = errors.New("grade não encontrada")
	ErrDiaBloqueado              = errors.New("dia bloqueado")
	ErrHorarioBloqueado          = errors.New("horário bloqueado")
	ErrHorarioIndisponivel       = errors.New("horário indisponível")
	ErrDataInvalida              = errors.New("data inválida")
	ErrHoraInvalida              = errors.New("hora inválida")
	ErrDadosInvalidos            = errors.New("dados inválidos")
	ErrProfissionalSemHorarios   = errors.New("profissional não trabalha com horários")
)

// Formato de data e hora gravado no banco (horário local, sem fuso).
// Datas e horas vão e voltam como texto para não depender do fuso da conexão.
const layoutDataHora = "2006-01-02 15:04:05"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) VerificarSalao(
	ctx context.Context,
	salaoID int64,
) error {
	const query = `
		SELECT id
		FROM saloes
		WHERE id = ?
		  AND status = 'ativo'
		LIMIT 1
	`

	var id int64

	err := r.db.QueryRowContext(
		ctx,
		query,
		salaoID,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrSalaoNaoEncontrado
	}

	return err
}

func (r *Repository) BuscarTextos(
	ctx context.Context,
	salaoID int64,
) (map[string]string, error) {
	const query = `
		SELECT chave, conteudo
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
		var chave string
		var conteudo string

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

// BuscarServicos lista todos os serviços ativos do salão.
// A tela de agendamento usa BuscarServicosDoProfissional.
func (r *Repository) BuscarServicos(
	ctx context.Context,
	salaoID int64,
) ([]Servico, error) {
	const query = `
		SELECT
			s.id,
			s.nome,
			s.descricao,
			s.img,
			ss.preco,
			ss.duracao_minutos
		FROM salao_servicos ss
		INNER JOIN servicos s
			ON s.id = ss.servico_id
		WHERE ss.salao_id = ?
		  AND ss.status = 'ativo'
		  AND s.status = 'ativo'
		ORDER BY s.nome ASC
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

	servicos := make([]Servico, 0)

	for rows.Next() {
		var servico Servico

		if err := rows.Scan(
			&servico.ID,
			&servico.Nome,
			&servico.Descricao,
			&servico.FotoURL,
			&servico.Preco,
			&servico.DuracaoMinutos,
		); err != nil {
			return nil, err
		}

		servicos = append(servicos, servico)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servicos, nil
}

// BuscarServicosDoProfissional lista só os serviços que o profissional faz
// (servico_profissional), com o preço e a duração do salão.
func (r *Repository) BuscarServicosDoProfissional(
	ctx context.Context,
	salaoID int64,
	profissionalID int64,
) ([]Servico, error) {
	const query = `
		SELECT
			s.id,
			s.nome,
			s.descricao,
			s.img,
			ss.preco,
			ss.duracao_minutos
		FROM servico_profissional sp
		INNER JOIN salao_servicos ss
			ON ss.servico_id = sp.servico_id
		   AND ss.salao_id = sp.salao_id
		   AND ss.status = 'ativo'
		INNER JOIN servicos s
			ON s.id = sp.servico_id
		   AND s.status = 'ativo'
		WHERE sp.salao_id = ?
		  AND sp.profissional_id = ?
		  AND sp.status = 'ativo'
		ORDER BY s.nome ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		salaoID,
		profissionalID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	servicos := make([]Servico, 0)

	for rows.Next() {
		var servico Servico

		if err := rows.Scan(
			&servico.ID,
			&servico.Nome,
			&servico.Descricao,
			&servico.FotoURL,
			&servico.Preco,
			&servico.DuracaoMinutos,
		); err != nil {
			return nil, err
		}

		servicos = append(servicos, servico)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servicos, nil
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
			foto_url,
			usa_horarios
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
		var usaHorarios int

		if err := rows.Scan(
			&profissional.ID,
			&profissional.Nome,
			&profissional.Especialidade,
			&profissional.FotoURL,
			&usaHorarios,
		); err != nil {
			return nil, err
		}

		profissional.UsaHorarios = usaHorarios == 1

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

func (r *Repository) BuscarProfissional(
	ctx context.Context,
	salaoID int64,
	profissionalID int64,
) (*Profissional, error) {
	const query = `
		SELECT
			id,
			nome,
			especialidade,
			foto_url,
			usa_horarios
		FROM profissionais
		WHERE id = ?
		  AND salao_id = ?
		  AND status = 'ativo'
		  AND deleted_at IS NULL
		LIMIT 1
	`

	var profissional Profissional
	var usaHorarios int

	err := r.db.QueryRowContext(
		ctx,
		query,
		profissionalID,
		salaoID,
	).Scan(
		&profissional.ID,
		&profissional.Nome,
		&profissional.Especialidade,
		&profissional.FotoURL,
		&usaHorarios,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProfissionalNaoEncontrado
	}

	if err != nil {
		return nil, err
	}

	profissional.UsaHorarios = usaHorarios == 1

	return &profissional, nil
}

func (r *Repository) BuscarGrade(
	ctx context.Context,
	salaoID int64,
	profissionalID int64,
	diaSemana string,
) (*Grade, error) {
	const query = `
		SELECT
			abre,
			abertura,
			pausa_inicio,
			pausa_fim,
			fechamento
		FROM profissional_grade_semanal
		WHERE salao_id = ?
		  AND profissional_id = ?
		  AND dia_semana = ?
		LIMIT 1
	`

	var grade Grade
	var abre int
	var abertura string
	var fechamento string
	var pausaInicio sql.NullString
	var pausaFim sql.NullString

	err := r.db.QueryRowContext(
		ctx,
		query,
		salaoID,
		profissionalID,
		diaSemana,
	).Scan(
		&abre,
		&abertura,
		&pausaInicio,
		&pausaFim,
		&fechamento,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrGradeNaoEncontrada
	}

	if err != nil {
		return nil, err
	}

	grade.Abre = abre == 1
	grade.Abertura = abertura
	grade.Fechamento = fechamento

	if pausaInicio.Valid {
		valor := pausaInicio.String
		grade.PausaInicio = &valor
	}

	if pausaFim.Valid {
		valor := pausaFim.String
		grade.PausaFim = &valor
	}

	return &grade, nil
}

func (r *Repository) VerificarDiaBloqueado(
	ctx context.Context,
	profissionalID int64,
	data string,
) (bool, error) {
	const query = `
		SELECT id
		FROM profissional_dias_bloqueados
		WHERE profissional_id = ?
		  AND data = ?
		LIMIT 1
	`

	var id int64

	err := r.db.QueryRowContext(
		ctx,
		query,
		profissionalID,
		data,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) VerificarHorarioBloqueado(
	ctx context.Context,
	profissionalID int64,
	inicio time.Time,
	fim time.Time,
) (bool, error) {
	const query = `
		SELECT id
		FROM profissional_horarios_bloqueados
		WHERE profissional_id = ?
		  AND inicio < ?
		  AND fim > ?
		LIMIT 1
	`

	var id int64

	// Texto no horário local, igual ao que fica gravado no banco.
	err := r.db.QueryRowContext(
		ctx,
		query,
		profissionalID,
		fim.Format(layoutDataHora),
		inicio.Format(layoutDataHora),
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// BuscarAgendamentosDoDia devolve os agendamentos ativos do profissional no dia,
// como intervalos: Inicio = hora_marcada, Fim = hora_marcada + duracao_minutos.
func (r *Repository) BuscarAgendamentosDoDia(
	ctx context.Context,
	salaoID int64,
	profissionalID int64,
	data string,
) ([]AgendamentoExistente, error) {
	const query = `
		SELECT
			a.id,
			DATE_FORMAT(a.hora_marcada, '%Y-%m-%d %H:%i:%s'),
			a.duracao_minutos
		FROM agendamentos a
		WHERE a.salao_id = ?
		  AND a.profissional_id = ?
		  AND a.hora_marcada >= ?
		  AND a.hora_marcada < DATE_ADD(?, INTERVAL 1 DAY)
		  AND a.status IN (
		      'pendente',
		      'confirmado'
		  )
		ORDER BY a.hora_marcada ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		salaoID,
		profissionalID,
		data,
		data,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	agendamentos := make([]AgendamentoExistente, 0)

	for rows.Next() {
		var (
			item        AgendamentoExistente
			horaMarcada string
			duracao     int
		)

		if err := rows.Scan(
			&item.ID,
			&horaMarcada,
			&duracao,
		); err != nil {
			return nil, err
		}

		inicio, err := time.ParseInLocation(
			layoutDataHora,
			horaMarcada,
			time.Local,
		)
		if err != nil {
			return nil, err
		}

		item.Inicio = inicio
		item.Fim = inicio.Add(
			time.Duration(duracao) * time.Minute,
		)

		agendamentos = append(
			agendamentos,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return agendamentos, nil
}

func (r *Repository) BuscarClienteAuthUserID(
	ctx context.Context,
	idCliente int64,
) (int64, error) {
	const query = `
		SELECT auth_user_id
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

func (r *Repository) BuscarServicosSelecionados(
	ctx context.Context,
	salaoID int64,
	profissionalID int64,
	servicoIDs []int64,
) ([]Servico, int, error) {
	if len(servicoIDs) == 0 {
		return nil, 0, ErrServicoNaoEncontrado
	}

	placeholders := make([]string, 0, len(servicoIDs))
	args := make([]any, 0, len(servicoIDs)+3)

	// A ordem dos argumentos segue a ordem dos "?" da query:
	// 1) profissional, 2) salão (servico_profissional), 3) ids, 4) salão (salao_servicos)
	args = append(args, profissionalID, salaoID)

	for _, id := range servicoIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	args = append(args, salaoID)

	query := `
		SELECT
			s.id,
			s.nome,
			s.descricao,
			ss.preco,
			ss.duracao_minutos
		FROM salao_servicos ss
		INNER JOIN servicos s
			ON s.id = ss.servico_id
		INNER JOIN servico_profissional sp
			ON sp.servico_id = ss.servico_id
		   AND sp.profissional_id = ?
		   AND sp.salao_id = ?
		   AND sp.status = 'ativo'
		WHERE ss.servico_id IN (` + strings.Join(placeholders, ",") + `)
		  AND ss.salao_id = ?
		  AND ss.status = 'ativo'
		  AND s.status = 'ativo'
		ORDER BY s.nome ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	servicos := make([]Servico, 0)
	duracaoTotal := 0

	for rows.Next() {
		var servico Servico

		if err := rows.Scan(
			&servico.ID,
			&servico.Nome,
			&servico.Descricao,
			&servico.Preco,
			&servico.DuracaoMinutos,
		); err != nil {
			return nil, 0, err
		}

		servicos = append(
			servicos,
			servico,
		)

		duracaoTotal += servico.DuracaoMinutos
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(servicos) != len(servicoIDs) {
		return nil, 0, ErrServicoNaoEncontrado
	}

	return servicos, duracaoTotal, nil
}

func (r *Repository) CriarAgendamento(
	ctx context.Context,
	req AgendarRequest,
	authUserID int64,
	servicos []Servico,
	horaMarcada time.Time,
	duracaoMinutos int,
) (int64, error) {
	tx, err := r.db.BeginTx(
		ctx,
		&sql.TxOptions{},
	)
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	// Confere novamente o profissional dentro da transação.
	const profissionalQuery = `
		SELECT id
		FROM profissionais
		WHERE id = ?
		  AND salao_id = ?
		  AND status = 'ativo'
		  AND deleted_at IS NULL
		LIMIT 1
	`

	var profissionalID int64

	err = tx.QueryRowContext(
		ctx,
		profissionalQuery,
		req.ProfissionalID,
		req.SalaoID,
	).Scan(&profissionalID)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrProfissionalNaoEncontrado
	}

	if err != nil {
		return 0, err
	}

	// Horário marcado e fim calculado, como texto (sem conversão de fuso).
	inicioTexto := horaMarcada.Format(layoutDataHora)

	fimTexto := horaMarcada.
		Add(time.Duration(duracaoMinutos) * time.Minute).
		Format(layoutDataHora)

	// Conflito real de intervalo:
	// existente = [hora_marcada, hora_marcada + duracao_minutos)
	const conflitoQuery = `
		SELECT id
		FROM agendamentos
		WHERE salao_id = ?
		  AND profissional_id = ?
		  AND status IN (
		      'pendente',
		      'confirmado'
		  )
		  AND hora_marcada < ?
		  AND DATE_ADD(hora_marcada, INTERVAL duracao_minutos MINUTE) > ?
		LIMIT 1
		FOR UPDATE
	`

	var conflitoID int64

	err = tx.QueryRowContext(
		ctx,
		conflitoQuery,
		req.SalaoID,
		req.ProfissionalID,
		fimTexto,
		inicioTexto,
	).Scan(&conflitoID)

	if err == nil {
		return 0, ErrHorarioIndisponivel
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	observacao := req.Observacao

	// hora_marcada    = data e hora escolhidas pelo cliente
	// duracao_minutos = soma da duração dos serviços escolhidos
	//
	// inicio/fim ainda são NOT NULL e outros módulos (agenda do profissional,
	// meus agendamentos, histórico) ainda leem essas colunas: por enquanto
	// recebem o mesmo intervalo, como espelho.
	const insertQuery = `
		INSERT INTO agendamentos (
			salao_id,
			profissional_id,
			auth_user_id,
			hora_marcada,
			duracao_minutos,
			inicio,
			fim,
			status,
			observacao
		)
		VALUES (
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			'pendente',
			?
		)
	`

	result, err := tx.ExecContext(
		ctx,
		insertQuery,
		req.SalaoID,
		req.ProfissionalID,
		authUserID,
		inicioTexto,
		duracaoMinutos,
		inicioTexto,
		fimTexto,
		observacao,
	)
	if err != nil {
		return 0, err
	}

	agendamentoID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	const snapshotQuery = `
		INSERT INTO servico_por_agendamento (
			agendamento_id,
			servico_id,
			nome_servico_snapshot,
			preco_unitario,
			duracao_minutos,
			quantidade
		)
		VALUES (
			?,
			?,
			?,
			?,
			?,
			1
		)
	`

	for _, servico := range servicos {
		_, err = tx.ExecContext(
			ctx,
			snapshotQuery,
			agendamentoID,
			servico.ID,
			servico.Nome,
			servico.Preco,
			servico.DuracaoMinutos,
		)

		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return agendamentoID, nil
}