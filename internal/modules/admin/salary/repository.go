package salary

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("profissional não encontrado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type ProfessionalSalary struct {
	Nome               string
	FotoURL            *string
	TipoRemuneracao    string
	PercentualComissao float64
	SalarioFixo        float64
}

func (r *Repository) BuscarProfissional(
	ctx context.Context,
	salaoID, profissionalID int64,
) (ProfessionalSalary, error) {
	var p ProfessionalSalary

	err := r.db.QueryRowContext(ctx, `
		SELECT
			nome,
			foto_url,
			tipo_remuneracao,
			percentual_comissao,
			salario_fixo
		FROM profissionais
		WHERE salao_id = ?
		  AND id = ?
		  AND deleted_at IS NULL
		  AND status = 'ativo'
		LIMIT 1
	`, salaoID, profissionalID).Scan(
		&p.Nome,
		&p.FotoURL,
		&p.TipoRemuneracao,
		&p.PercentualComissao,
		&p.SalarioFixo,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return ProfessionalSalary{}, ErrNotFound
	}

	return p, err
}

func (r *Repository) BuscarRendimentos(
	ctx context.Context,
	salaoID, profissionalID int64,
	inicio, fim time.Time,
) ([]AppointmentEarning, error) {
	query := `
		SELECT
			a.id,
			DATE_FORMAT(a.inicio, '%Y-%m-%d') AS data,
			COALESCE(
				SUM(spa.preco_unitario * spa.quantidade),
				0
			) AS total
		FROM agendamentos a
		INNER JOIN servico_por_agendamento spa
			ON spa.agendamento_id = a.id
		WHERE a.salao_id = ?
		  AND a.profissional_id = ?
		  AND a.status = 'concluido'
		  AND a.inicio >= ?
		  AND a.inicio < ?
		GROUP BY a.id, a.inicio
		ORDER BY a.inicio ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		salaoID,
		profissionalID,
		inicio,
		fim,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]AppointmentEarning, 0)

	for rows.Next() {
		var item AppointmentEarning

		if err := rows.Scan(
			&item.ID,
			&item.Data,
			&item.Total,
		); err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
