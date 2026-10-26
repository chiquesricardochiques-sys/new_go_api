package statusAG

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNaoEncontrado = errors.New("registro não encontrado")

type Repository struct {
	db *sql.DB
}

const layoutSQL = "2006-01-02 15:04:05"

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

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

func (r *Repository) IniciarCorteRepo(ctx context.Context, idSalao, profissionalID, agendamentoID int64, horaInicio time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE agendamentos
		SET status = 'confirmado',
		    inicio = ?
		WHERE id = ? AND salao_id = ? AND profissional_id = ?`,
		horaInicio.Format(layoutSQL), agendamentoID, idSalao, profissionalID,
	)
	return err
}

func (r *Repository) FinalizarCorteRepo(ctx context.Context, idSalao, profissionalID, agendamentoID int64, horaFim time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE agendamentos a
		JOIN profissionais p ON p.id = a.profissional_id
		SET a.status = 'concluido',
		    a.fim = ?,
		    a.percentual_comissao = COALESCE(a.percentual_comissao, p.percentual_comissao)
		WHERE a.id = ? AND a.salao_id = ? AND a.profissional_id = ?`,
		horaFim.Format(layoutSQL), agendamentoID, idSalao, profissionalID,
	)
	return err
}

func (r *Repository) CancelarAgendamentoRepo(ctx context.Context, idSalao, profissionalID, agendamentoID int64, observacaoExtra string) error {
	query := `UPDATE agendamentos SET status = 'cancelado'`
	args := []any{}
	if observacaoExtra != "" {
		query += `, observacao = CONCAT(COALESCE(observacao, ''), '\n[Cancelado]: ?')`
		args = append(args, observacaoExtra)
	}
	query += ` WHERE id = ? AND salao_id = ? AND profissional_id = ?`
	args = append(args, agendamentoID, idSalao, profissionalID)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) NaoCompareceuRepo(ctx context.Context, idSalao, profissionalID, agendamentoID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE agendamentos
		SET status = 'nao_compareceu'
		WHERE id = ? AND salao_id = ? AND profissional_id = ?`,
		agendamentoID, idSalao, profissionalID,
	)
	return err
}