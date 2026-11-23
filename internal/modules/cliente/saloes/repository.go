package saloes

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

func escaparLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// Buscar lista salões ativos. A ordem dos args segue a ordem dos "?" no SQL.
func (r *Repository) Buscar(
	ctx context.Context,
	f filtro,
) ([]SalaoItem, error) {
	var sb strings.Builder
	args := make([]any, 0, 16)

	temLoc := f.Latitude != nil && f.Longitude != nil

	sb.WriteString(`
		SELECT
			s.id, s.nome, s.slug, s.endereco, s.cidade, s.estado,
			s.whatsapp, s.logo_url, s.img_salao,`)

	if temLoc {
		// ST_Distance_Sphere usa POINT(longitude, latitude) e devolve metros
		sb.WriteString(`
			CASE
				WHEN s.latitude IS NULL OR s.longitude IS NULL THEN NULL
				ELSE ST_Distance_Sphere(POINT(s.longitude, s.latitude), POINT(?, ?)) / 1000
			END AS distancia_km,`)
		args = append(args, *f.Longitude, *f.Latitude)
	} else {
		sb.WriteString(` NULL AS distancia_km,`)
	}

	// aberto: NULL = salão sem horário cadastrado (não dá para afirmar nada)
	sb.WriteString(`
			(
				SELECT CASE
					WHEN COUNT(*) = 0 THEN NULL
					WHEN SUM(h.dia_semana = ? AND h.abre = 1
					         AND h.hora_abertura <= ? AND h.hora_fechamento > ?) > 0 THEN 1
					ELSE 0
				END
				FROM horario_funcionamento h
				WHERE h.salao_id = s.id
			) AS aberto,`)
	args = append(args, f.Dia, f.Hora, f.Hora)

	sb.WriteString(`
			av.nota_media,
			COALESCE(av.total, 0) AS total_avaliacoes,
			(
				SELECT MIN(ss.preco) FROM salao_servicos ss
				WHERE ss.salao_id = s.id AND ss.status = 'ativo'
			) AS preco_inicial,
			(
				SELECT COUNT(*) FROM profissionais p
				WHERE p.salao_id = s.id AND p.status = 'ativo'
			) AS total_profissionais,
			(
				SELECT GROUP_CONCAT(DISTINCT sv.nome ORDER BY sv.nome SEPARATOR '|')
				FROM salao_servicos ss
				JOIN servicos sv ON sv.id = ss.servico_id
				WHERE ss.salao_id = s.id AND ss.status = 'ativo' AND sv.status = 'ativo'
			) AS especialidades,`)

	if f.AuthUserID > 0 {
		sb.WriteString(`
			EXISTS (
				SELECT 1 FROM agendamentos a
				WHERE a.salao_id = s.id
				  AND a.auth_user_id = ?
				  AND a.status IN ('pendente', 'confirmado')
				  AND a.hora_marcada >= NOW()
			) AS tem_agendamento`)
		args = append(args, f.AuthUserID)
	} else {
		sb.WriteString(` 0 AS tem_agendamento`)
	}

	sb.WriteString(`
		FROM saloes s
		LEFT JOIN (
			SELECT salao_id, AVG(nota) AS nota_media, COUNT(*) AS total
			FROM avaliacoes
			GROUP BY salao_id
		) av ON av.salao_id = s.id
		WHERE s.status = 'ativo'`)

	if f.Busca != "" {
		like := "%" + escaparLike(f.Busca) + "%"
		sb.WriteString(` AND (s.nome LIKE ? OR s.cidade LIKE ? OR s.endereco LIKE ?)`)
		args = append(args, like, like, like)
	}

	if f.Cidade != "" {
		sb.WriteString(` AND s.cidade = ?`)
		args = append(args, f.Cidade)
	}

	if f.Estado != "" {
		sb.WriteString(` AND s.estado = ?`)
		args = append(args, f.Estado)
	}

	if temLoc {
		sb.WriteString(` ORDER BY distancia_km IS NULL, distancia_km ASC, s.nome ASC`)
	} else {
		sb.WriteString(` ORDER BY s.nome ASC`)
	}

	sb.WriteString(` LIMIT ? OFFSET ?`)
	args = append(args, f.Limite, f.Offset)

	rows, err := r.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itens := make([]SalaoItem, 0)

	for rows.Next() {
		var (
			item        SalaoItem
			distancia   sql.NullFloat64
			aberto      sql.NullInt64
			nota        sql.NullFloat64
			preco       sql.NullFloat64
			esp         sql.NullString
			temAgendInt int
		)

		if err := rows.Scan(
			&item.ID,
			&item.Nome,
			&item.Slug,
			&item.Endereco,
			&item.Cidade,
			&item.Estado,
			&item.Whatsapp,
			&item.LogoURL,
			&item.ImgSalao,
			&distancia,
			&aberto,
			&nota,
			&item.TotalAvaliacoes,
			&preco,
			&item.TotalProfissionais,
			&esp,
			&temAgendInt,
		); err != nil {
			return nil, err
		}

		if distancia.Valid {
			v := distancia.Float64
			item.DistanciaKm = &v
		}

		if aberto.Valid {
			v := aberto.Int64 == 1
			item.Aberto = &v
		}

		if nota.Valid {
			v := nota.Float64
			item.NotaMedia = &v
		}

		if preco.Valid {
			v := preco.Float64
			item.PrecoInicial = &v
		}

		item.Especialidades = []string{}
		if esp.Valid && esp.String != "" {
			partes := strings.Split(esp.String, "|")
			if len(partes) > 3 {
				partes = partes[:3]
			}
			item.Especialidades = partes
		}

		item.TemAgendamento = temAgendInt == 1

		itens = append(itens, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return itens, nil
}

func (r *Repository) BuscarPorSlug(
	ctx context.Context,
	slug string,
) (*SalaoItem, error) {
	const query = `
		SELECT id, nome, slug, endereco, cidade, estado, whatsapp, logo_url, img_salao
		FROM saloes
		WHERE slug = ?
		  AND status = 'ativo'
		LIMIT 1
	`

	var item SalaoItem

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&item.ID, &item.Nome, &item.Slug, &item.Endereco,
		&item.Cidade, &item.Estado, &item.Whatsapp, &item.LogoURL, &item.ImgSalao,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSalaoNaoEncontrado
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
}

// Detalhes monta a página pública do salão (serviços, equipe, horários, avaliações).
func (r *Repository) Detalhes(
	ctx context.Context,
	slug string,
) (*DetalhesData, error) {
	salao, err := r.BuscarPorSlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	d := &DetalhesData{
		Salao:         *salao,
		Servicos:      []ServicoItem{},
		Profissionais: []ProfissionalItem{},
		Horarios:      []HorarioItem{},
		Avaliacoes:    []AvaliacaoItem{},
	}

	// nota média + total
	var nota sql.NullFloat64
	if err := r.db.QueryRowContext(ctx,
		`SELECT AVG(nota), COUNT(*) FROM avaliacoes WHERE salao_id = ?`, salao.ID,
	).Scan(&nota, &d.Salao.TotalAvaliacoes); err != nil {
		return nil, err
	}
	if nota.Valid {
		v := nota.Float64
		d.Salao.NotaMedia = &v
	}

	// serviços
	rs, err := r.db.QueryContext(ctx, `
		SELECT sv.id, sv.nome, sv.descricao, sv.img, ss.preco, ss.duracao_minutos
		FROM salao_servicos ss
		JOIN servicos sv ON sv.id = ss.servico_id
		WHERE ss.salao_id = ? AND ss.status = 'ativo' AND sv.status = 'ativo'
		ORDER BY sv.nome`, salao.ID)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	for rs.Next() {
		var s ServicoItem
		if err := rs.Scan(&s.ID, &s.Nome, &s.Descricao, &s.Img, &s.Preco, &s.DuracaoMin); err != nil {
			return nil, err
		}
		d.Servicos = append(d.Servicos, s)
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}

	// profissionais (ajuste os nomes das colunas se forem diferentes)
	rp, err := r.db.QueryContext(ctx, `
		SELECT id, nome, especialidade, foto_url
		FROM profissionais
		WHERE salao_id = ? AND status = 'ativo'
		ORDER BY nome`, salao.ID)
	if err != nil {
		return nil, err
	}
	defer rp.Close()

	for rp.Next() {
		var p ProfissionalItem
		if err := rp.Scan(&p.ID, &p.Nome, &p.Especialidade, &p.FotoURL); err != nil {
			return nil, err
		}
		d.Profissionais = append(d.Profissionais, p)
	}
	if err := rp.Err(); err != nil {
		return nil, err
	}

	// horário de funcionamento
	rh, err := r.db.QueryContext(ctx, `
		SELECT dia_semana, abre,
		       TIME_FORMAT(hora_abertura, '%H:%i'),
		       TIME_FORMAT(hora_fechamento, '%H:%i')
		FROM horario_funcionamento
		WHERE salao_id = ?
		ORDER BY FIELD(dia_semana, 'seg','ter','qua','qui','sex','sab','dom')`, salao.ID)
	if err != nil {
		return nil, err
	}
	defer rh.Close()

	for rh.Next() {
		var h HorarioItem
		if err := rh.Scan(&h.Dia, &h.Abre, &h.Abertura, &h.Fechamento); err != nil {
			return nil, err
		}
		d.Horarios = append(d.Horarios, h)
	}
	if err := rh.Err(); err != nil {
		return nil, err
	}

	// últimas avaliações
	ra, err := r.db.QueryContext(ctx, `
		SELECT nota, comentario, DATE_FORMAT(created_at, '%d/%m/%Y')
		FROM avaliacoes
		WHERE salao_id = ?
		ORDER BY created_at DESC
		LIMIT 5`, salao.ID)
	if err != nil {
		return nil, err
	}
	defer ra.Close()

	for ra.Next() {
		var a AvaliacaoItem
		if err := ra.Scan(&a.Nota, &a.Comentario, &a.Data); err != nil {
			return nil, err
		}
		d.Avaliacoes = append(d.Avaliacoes, a)
	}

	return d, ra.Err()
}

// Avaliar grava a avaliação só se o agendamento é do cliente, está concluído
// e ainda não foi avaliado. Tudo numa única query (sem corrida entre checar e gravar).
func (r *Repository) Avaliar(
	ctx context.Context,
	req AvaliarRequest,
) error {
	var comentario any
	if req.Comentario != "" {
		comentario = req.Comentario
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO avaliacoes
			(salao_id, profissional_id, agendamento_id, auth_user_id, nota, comentario)
		SELECT a.salao_id, a.profissional_id, a.id, a.auth_user_id, ?, ?
		FROM agendamentos a
		WHERE a.id = ?
		  AND a.auth_user_id = ?
		  AND a.status = 'concluido'
		  AND NOT EXISTS (SELECT 1 FROM avaliacoes v WHERE v.agendamento_id = a.id)`,
		req.Nota, comentario, req.AgendamentoID, req.AuthUserID,
	)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrNaoPodeAvaliar
	}

	return nil
}

func (r *Repository) ListarCidades(ctx context.Context) ([]Cidade, error) {
	const query = `
		SELECT cidade, estado, COUNT(*)
		FROM saloes
		WHERE status = 'ativo'
		  AND cidade IS NOT NULL AND cidade <> ''
		  AND estado IS NOT NULL AND estado <> ''
		GROUP BY estado, cidade
		ORDER BY estado ASC, cidade ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cidades := make([]Cidade, 0)

	for rows.Next() {
		var c Cidade

		if err := rows.Scan(&c.Cidade, &c.Estado, &c.Total); err != nil {
			return nil, err
		}

		cidades = append(cidades, c)
	}

	return cidades, rows.Err()
}