
package professionals

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("profissional não encontrado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const selectFields = `
	id,
	salao_id,
	nome,
	especialidade,
	foto_url,
	foto_public_id,
	status,
	usa_horarios,
	tipo_remuneracao,
	percentual_comissao,
	salario_fixo,
	frequencia_pagamento,
	created_at,
	updated_at
`

func scanProfessional(scanner interface {
	Scan(dest ...any) error
}) (Professional, error) {
	var p Professional
	var usaHorarios int

	err := scanner.Scan(
		&p.ID,
		&p.SalaoID,
		&p.Nome,
		&p.Especialidade,
		&p.FotoURL,
		&p.FotoPublicID,
		&p.Status,
		&usaHorarios,
		&p.TipoRemuneracao,
		&p.PercentualComissao,
		&p.SalarioFixo,
		&p.FrequenciaPagamento,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return p, err
	}

	p.UsaHorarios = usaHorarios != 0
	return p, nil
}

// Listar retorna todos os profissionais do salão.
func (r *Repository) Listar(
	ctx context.Context,
	salaoID int64,
) ([]Professional, error) {
	query := `
		SELECT ` + selectFields + `
		FROM profissionais
		WHERE salao_id = ?
		  AND deleted_at IS NULL
		ORDER BY nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query, salaoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profissionais := make([]Professional, 0)

	for rows.Next() {
		p, err := scanProfessional(rows)
		if err != nil {
			return nil, err
		}

		profissionais = append(profissionais, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profissionais, nil
}

// Buscar retorna um profissional específico dentro do salão.
func (r *Repository) Buscar(
	ctx context.Context,
	salaoID int64,
	id int64,
) (Professional, error) {
	query := `
		SELECT ` + selectFields + `
		FROM profissionais
		WHERE salao_id = ?
		  AND id = ?
		  AND deleted_at IS NULL
		LIMIT 1
	`

	p, err := scanProfessional(
		r.db.QueryRowContext(ctx, query, salaoID, id),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Professional{}, ErrNotFound
	}

	return p, err
}

// Atualizar modifica somente os campos permitidos.
func (r *Repository) Atualizar(
	ctx context.Context,
	req *UpdateRequest,
) error {
	current, err := r.Buscar(ctx, req.SalaoID, req.ID)
	if err != nil {
		return err
	}

	if req.Nome != nil {
		current.Nome = *req.Nome
	}

	if req.Especialidade != nil {
		value := *req.Especialidade
		if value == "" {
			current.Especialidade = nil
		} else {
			current.Especialidade = &value
		}
	}

	if req.Status != nil {
		current.Status = *req.Status
	}

	if req.UsaHorarios != nil {
		current.UsaHorarios = *req.UsaHorarios
	}

	if req.TipoRemuneracao != nil {
		current.TipoRemuneracao = *req.TipoRemuneracao
	}

	if req.PercentualComissao != nil {
		current.PercentualComissao = *req.PercentualComissao
	}

	if req.SalarioFixo != nil {
		current.SalarioFixo = *req.SalarioFixo
	}

	if req.FrequenciaPagamento != nil {
		current.FrequenciaPagamento = *req.FrequenciaPagamento
	}

	query := `
		UPDATE profissionais
		SET nome = ?,
			especialidade = ?,
			status = ?,
			usa_horarios = ?,
			tipo_remuneracao = ?,
			percentual_comissao = ?,
			salario_fixo = ?,
			frequencia_pagamento = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE salao_id = ?
		  AND id = ?
		  AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		current.Nome,
		current.Especialidade,
		current.Status,
		current.UsaHorarios,
		current.TipoRemuneracao,
		current.PercentualComissao,
		current.SalarioFixo,
		current.FrequenciaPagamento,
		req.SalaoID,
		req.ID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		// O registro pode continuar existindo mesmo quando
		// nenhum valor foi alterado.
		_, err = r.Buscar(ctx, req.SalaoID, req.ID)
		return err
	}

	return nil
}
