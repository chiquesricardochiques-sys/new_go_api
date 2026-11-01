package verAgendamentosAvulso

import (
    "context"
    "database/sql"
    "errors"
    "time"
)

var (
    ErrAgendamentoNaoEncontrado = errors.New("agendamento avulso não encontrado")
    ErrAgendamentoJaCancelado   = errors.New("agendamento já cancelado ou não pode ser cancelado")
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

// Listar retorna somente agendamentos avulsos do salão e do dia informado.
func (r *Repository) Listar(
    ctx context.Context,
    filtro FiltroAgendamentosAvulsos,
) ([]AgendamentoAvulso, error) {

    data, err := time.Parse("2006-01-02", filtro.Data)
    if err != nil {
        return nil, errors.New("data inválida; use YYYY-MM-DD")
    }

    inicio := data.Format("2006-01-02 00:00:00")
    fim := data.AddDate(0, 0, 1).Format("2006-01-02 00:00:00")

    query := `
        SELECT
            a.id,
            DATE_FORMAT(a.inicio, '%Y-%m-%d') AS data,
            DATE_FORMAT(a.inicio, '%H:%i') AS hora,
            DATE_FORMAT(a.fim, '%H:%i') AS hora_fim,
            COALESCE(a.nome_cliente, '') AS nome_cliente,
            COALESCE(a.telefone_cliente, '') AS telefone_cliente,
            a.duracao_minutos,
            a.status,
            COALESCE(a.observacao, '') AS observacao,
            CASE
                WHEN a.auth_user_id IS NULL THEN 'avulso_admin'
                ELSE 'site'
            END AS origem
        FROM agendamentos a
        WHERE
            a.salao_id = ?
            AND a.auth_user_id IS NULL
            AND a.inicio >= ?
            AND a.inicio < ?
        ORDER BY a.inicio ASC, a.id ASC
    `

    rows, err := r.db.QueryContext(
        ctx,
        query,
        filtro.SalaoID,
        inicio,
        fim,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    agendamentos := make([]AgendamentoAvulso, 0)

    for rows.Next() {
        var item AgendamentoAvulso

        err := rows.Scan(
            &item.ID,
            &item.Data,
            &item.Hora,
            &item.HoraFim,
            &item.NomeCliente,
            &item.TelefoneCliente,
            &item.DuracaoMinutos,
            &item.Status,
            &item.Observacao,
            &item.Origem,
        )
        if err != nil {
            return nil, err
        }

        agendamentos = append(agendamentos, item)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return agendamentos, nil
}

// Cancelar altera somente agendamentos avulsos do salão informado.
func (r *Repository) Cancelar(
    ctx context.Context,
    salaID int64,
    agendamentoID int64,
) error {

    query := `
        UPDATE agendamentos
        SET
            status = 'cancelado',
            updated_at = CURRENT_TIMESTAMP
        WHERE
            id = ?
            AND salao_id = ?
            AND auth_user_id IS NULL
            AND status IN ('pendente', 'confirmado')
    `

    result, err := r.db.ExecContext(
        ctx,
        query,
        agendamentoID,
        salaID,
    )
    if err != nil {
        return err
    }

    afetadas, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if afetadas == 0 {
        // Verifica se o registro existe dentro do salão.
        var existe bool

        err := r.db.QueryRowContext(ctx, `
            SELECT EXISTS (
                SELECT 1
                FROM agendamentos
                WHERE id = ?
                  AND salao_id = ?
                  AND auth_user_id IS NULL
            )
        `, agendamentoID, salaID).Scan(&existe)

        if err != nil {
            return err
        }

        if !existe {
            return ErrAgendamentoNaoEncontrado
        }

        return ErrAgendamentoJaCancelado
    }

    return nil
}