package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// ============================================================
// SERVIÇOS GLOBAIS
// ============================================================

func (r *Repository) ListGlobalServices(
	ctx context.Context,
) ([]Servico, error) {

	const query = `
		SELECT
			id,
			nome,
			descricao,
			status,
			created_at,
			updated_at
		FROM servicos
		WHERE status = 'ativo'
		ORDER BY nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
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
			&servico.Status,
			&servico.CreatedAt,
			&servico.UpdatedAt,
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

// ============================================================
// SERVIÇOS DO SALÃO
// ============================================================

func (r *Repository) ListSalonServices(
	ctx context.Context,
	salaoID uint64,
) ([]ServicoSalao, error) {

	const query = `
		SELECT
			s.id,
			COALESCE(ss.id, 0) AS salao_servico_id,
			s.nome,
			s.descricao,
			COALESCE(ss.preco, 0) AS preco,
			COALESCE(ss.duracao_minutos, 0) AS duracao_minutos,
			COALESCE(ss.status, 'inativo') AS status,
			CASE
				WHEN ss.id IS NOT NULL
				 AND ss.status = 'ativo'
				THEN 1
				ELSE 0
			END AS selecionado,
			COALESCE(ss.created_at, s.created_at) AS created_at,
			COALESCE(ss.updated_at, s.updated_at) AS updated_at
		FROM servicos s
		LEFT JOIN salao_servicos ss
			ON ss.servico_id = s.id
			AND ss.salao_id = ?
		WHERE s.status = 'ativo'
		ORDER BY s.nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query, salaoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servicos := make([]ServicoSalao, 0)

	for rows.Next() {
		var item ServicoSalao
		var selecionado int

		if err := rows.Scan(
			&item.ID,
			&item.SalaoServicoID,
			&item.Nome,
			&item.Descricao,
			&item.Preco,
			&item.DuracaoMinutos,
			&item.Status,
			&selecionado,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		item.Selecionado = selecionado == 1

		servicos = append(servicos, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servicos, nil
}

func (r *Repository) GetSalonService(
	ctx context.Context,
	salaoID uint64,
	servicoID uint64,
) (*ServicoSalao, error) {

	const query = `
		SELECT
			s.id,
			ss.id,
			s.nome,
			s.descricao,
			ss.preco,
			ss.duracao_minutos,
			ss.status,
			1 AS selecionado,
			ss.created_at,
			ss.updated_at
		FROM servicos s
		INNER JOIN salao_servicos ss
			ON ss.servico_id = s.id
		WHERE ss.salao_id = ?
			AND ss.servico_id = ?
			AND ss.status = 'ativo'
			AND s.status = 'ativo'
		LIMIT 1
	`

	var item ServicoSalao
	var selecionado int

	err := r.db.QueryRowContext(
		ctx,
		query,
		salaoID,
		servicoID,
	).Scan(
		&item.ID,
		&item.SalaoServicoID,
		&item.Nome,
		&item.Descricao,
		&item.Preco,
		&item.DuracaoMinutos,
		&item.Status,
		&selecionado,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, err
	}

	item.Selecionado = selecionado == 1

	return &item, nil
}

func (r *Repository) AddSalonService(
	ctx context.Context,
	salaoID uint64,
	servicoID uint64,
	preco float64,
	duracaoMinutos uint,
) error {

	const query = `
		INSERT INTO salao_servicos (
			salao_id,
			servico_id,
			preco,
			duracao_minutos,
			status,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, 'ativo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON DUPLICATE KEY UPDATE
			preco = VALUES(preco),
			duracao_minutos = VALUES(duracao_minutos),
			status = 'ativo',
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		salaoID,
		servicoID,
		preco,
		duracaoMinutos,
	)

	return err
}

func (r *Repository) DeactivateSalonService(
	ctx context.Context,
	salaoID uint64,
	servicoID uint64,
) error {

	const query = `
		UPDATE salao_servicos
		SET
			status = 'inativo',
			updated_at = CURRENT_TIMESTAMP
		WHERE salao_id = ?
			AND servico_id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		salaoID,
		servicoID,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repository) UnlinkSalonService(
	ctx context.Context,
	salaoID uint64,
	servicoID uint64,
) error {

	const query = `
		DELETE FROM salao_servicos
		WHERE salao_id = ?
			AND servico_id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		salaoID,
		servicoID,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ============================================================
// PROFISSIONAIS
// ============================================================

func (r *Repository) ListProfessionals(
	ctx context.Context,
	salaoID uint64,
) ([]Profissional, error) {

	const query = `
		SELECT
			id,
			nome,
			especialidade,
			status
		FROM profissionais
		WHERE salao_id = ?
			AND status = 'ativo'
			AND deleted_at IS NULL
		ORDER BY nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query, salaoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profissionais := make([]Profissional, 0)

	for rows.Next() {
		var profissional Profissional

		if err := rows.Scan(
			&profissional.ID,
			&profissional.Nome,
			&profissional.Especialidade,
			&profissional.Status,
		); err != nil {
			return nil, err
		}

		profissionais = append(profissionais, profissional)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profissionais, nil
}

func (r *Repository) ListProfessionalsByService(
	ctx context.Context,
	salaoID uint64,
	servicoID uint64,
) ([]Profissional, error) {

	const query = `
		SELECT
			p.id,
			p.nome,
			p.especialidade,
			p.status
		FROM profissionais p
		INNER JOIN servico_profissional sp
			ON sp.profissional_id = p.id
			AND sp.salao_id = p.salao_id
			AND sp.servico_id = ?
			AND sp.status = 'ativo'
		INNER JOIN salao_servicos ss
			ON ss.salao_id = p.salao_id
			AND ss.servico_id = ?
			AND ss.status = 'ativo'
		INNER JOIN servicos s
			ON s.id = ss.servico_id
			AND s.status = 'ativo'
		WHERE p.salao_id = ?
			AND p.status = 'ativo'
			AND p.deleted_at IS NULL
		ORDER BY p.nome ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		servicoID,
		servicoID,
		salaoID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profissionais := make([]Profissional, 0)

	for rows.Next() {
		var profissional Profissional

		if err := rows.Scan(
			&profissional.ID,
			&profissional.Nome,
			&profissional.Especialidade,
			&profissional.Status,
		); err != nil {
			return nil, err
		}

		profissionais = append(profissionais, profissional)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profissionais, nil
}

// ============================================================
// SERVIÇOS DO PROFISSIONAL
// ============================================================

// Checklist:
// retorna todos os serviços ativos do salão,
// marcando Selecionado=true quando o profissional possui o vínculo.
func (r *Repository) ListProfessionalServices(
	ctx context.Context,
	salaoID uint64,
	profissionalID uint64,
) ([]ServicoProfissionalItem, error) {

	const query = `
		SELECT
			s.id,
			ss.id,
			s.nome,
			s.descricao,
			ss.preco,
			ss.duracao_minutos,
			ss.status,
			CASE
				WHEN sp.id IS NOT NULL
				 AND sp.status = 'ativo'
				THEN 1
				ELSE 0
			END AS selecionado
		FROM salao_servicos ss
		INNER JOIN servicos s
			ON s.id = ss.servico_id
			AND s.status = 'ativo'
		LEFT JOIN servico_profissional sp
			ON sp.salao_id = ss.salao_id
			AND sp.profissional_id = ?
			AND sp.servico_id = ss.servico_id
		WHERE ss.salao_id = ?
			AND ss.status = 'ativo'
		ORDER BY s.nome ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		profissionalID,
		salaoID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servicos := make([]ServicoProfissionalItem, 0)

	for rows.Next() {
		var item ServicoProfissionalItem
		var selecionado int

		if err := rows.Scan(
			&item.ID,
			&item.SalaoServicoID,
			&item.Nome,
			&item.Descricao,
			&item.Preco,
			&item.DuracaoMinutos,
			&item.Status,
			&selecionado,
		); err != nil {
			return nil, err
		}

		item.Selecionado = selecionado == 1

		servicos = append(servicos, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servicos, nil
}

// Somente os serviços efetivamente vinculados.
func (r *Repository) ListLinkedProfessionalServices(
	ctx context.Context,
	salaoID uint64,
	profissionalID uint64,
) ([]ServicoProfissionalItem, error) {

	const query = `
		SELECT
			s.id,
			ss.id,
			s.nome,
			s.descricao,
			ss.preco,
			ss.duracao_minutos,
			ss.status,
			1 AS selecionado
		FROM servico_profissional sp
		INNER JOIN salao_servicos ss
			ON ss.salao_id = sp.salao_id
			AND ss.servico_id = sp.servico_id
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

	servicos := make([]ServicoProfissionalItem, 0)

	for rows.Next() {
		var item ServicoProfissionalItem
		var selecionado int

		if err := rows.Scan(
			&item.ID,
			&item.SalaoServicoID,
			&item.Nome,
			&item.Descricao,
			&item.Preco,
			&item.DuracaoMinutos,
			&item.Status,
			&selecionado,
		); err != nil {
			return nil, err
		}

		item.Selecionado = selecionado == 1

		servicos = append(servicos, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servicos, nil
}

func (r *Repository) AddProfessionalService(
	ctx context.Context,
	salaoID uint64,
	profissionalID uint64,
	servicoID uint64,
) error {

	const query = `
		INSERT INTO servico_profissional (
			salao_id,
			profissional_id,
			servico_id,
			status,
			created_at
		)
		VALUES (?, ?, ?, 'ativo', CURRENT_TIMESTAMP)
		ON DUPLICATE KEY UPDATE
			salao_id = VALUES(salao_id),
			status = 'ativo'
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		salaoID,
		profissionalID,
		servicoID,
	)

	return err
}

func (r *Repository) RemoveProfessionalService(
	ctx context.Context,
	salaoID uint64,
	profissionalID uint64,
	servicoID uint64,
) error {

	const query = `
		UPDATE servico_profissional
		SET status = 'inativo'
		WHERE salao_id = ?
			AND profissional_id = ?
			AND servico_id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		salaoID,
		profissionalID,
		servicoID,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ============================================================
// VALIDAÇÕES
// ============================================================

func (r *Repository) ValidateGlobalService(
	ctx context.Context,
	servicoID uint64,
) error {

	const query = `
		SELECT COUNT(*)
		FROM servicos
		WHERE id = ?
			AND status = 'ativo'
	`

	var count int

	if err := r.db.QueryRowContext(
		ctx,
		query,
		servicoID,
	).Scan(&count); err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("serviço global não encontrado ou inativo")
	}

	return nil
}

func (r *Repository) ValidateSalonService(
	ctx context.Context,
	salaoID uint64,
	servicoID uint64,
) error {

	const query = `
		SELECT COUNT(*)
		FROM salao_servicos ss
		INNER JOIN servicos s
			ON s.id = ss.servico_id
		WHERE ss.salao_id = ?
			AND ss.servico_id = ?
			AND ss.status = 'ativo'
			AND s.status = 'ativo'
	`

	var count int

	if err := r.db.QueryRowContext(
		ctx,
		query,
		salaoID,
		servicoID,
	).Scan(&count); err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("serviço não está ativo neste salão")
	}

	return nil
}

func (r *Repository) ValidateSalonServiceRelation(
	ctx context.Context,
	salaoID uint64,
	servicoID uint64,
) error {

	const query = `
		SELECT COUNT(*)
		FROM salao_servicos
		WHERE salao_id = ?
			AND servico_id = ?
	`

	var count int

	if err := r.db.QueryRowContext(
		ctx,
		query,
		salaoID,
		servicoID,
	).Scan(&count); err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("serviço não está vinculado ao salão")
	}

	return nil
}

func (r *Repository) ValidateProfessional(
	ctx context.Context,
	salaoID uint64,
	profissionalID uint64,
) error {

	const query = `
		SELECT COUNT(*)
		FROM profissionais
		WHERE id = ?
			AND salao_id = ?
			AND status = 'ativo'
			AND deleted_at IS NULL
	`

	var count int

	if err := r.db.QueryRowContext(
		ctx,
		query,
		profissionalID,
		salaoID,
	).Scan(&count); err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("profissional não encontrado ou inativo")
	}

	return nil
}