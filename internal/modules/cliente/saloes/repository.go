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
	args := make([]any, 0, 12)

	temLoc := f.Latitude != nil && f.Longitude != nil

	sb.WriteString(`
		SELECT
			s.id, s.nome, s.slug, s.endereco, s.cidade, s.estado,
			s.whatsapp, s.logo_url,`)

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

	sb.WriteString(` FROM saloes s WHERE s.status = 'ativo'`)

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
			temAgendInt int
		)

		if err := rows.Scan(
			&item.ID, &item.Nome, &item.Slug, &item.Endereco,
			&item.Cidade, &item.Estado, &item.Whatsapp, &item.LogoURL,
			&distancia, &temAgendInt,
		); err != nil {
			return nil, err
		}

		if distancia.Valid {
			v := distancia.Float64
			item.DistanciaKm = &v
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
		SELECT id, nome, slug, endereco, cidade, estado, whatsapp, logo_url
		FROM saloes
		WHERE slug = ?
		  AND status = 'ativo'
		LIMIT 1
	`

	var item SalaoItem

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&item.ID, &item.Nome, &item.Slug, &item.Endereco,
		&item.Cidade, &item.Estado, &item.Whatsapp, &item.LogoURL,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSalaoNaoEncontrado
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
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