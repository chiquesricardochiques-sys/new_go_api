package dadosagendamentos

import (
	"context"
	"database/sql"
	"errors"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// BuscarAgendamento devolve o cabeçalho do agendamento (com profissional e salão).
//
// O JOIN com clientes_auth_vinculos garante que o agendamento pertence ao
// cliente informado: se for de outro cliente, a consulta não retorna linha.
func (r *Repository) BuscarAgendamento(
	ctx context.Context,
	req *DadosRequest,
) (*AgendamentoDados, error) {

	const query = `
		SELECT
			a.id,
			a.status,
			DATE_FORMAT(a.hora_marcada, '%Y-%m-%d'),
			TIME_FORMAT(a.hora_marcada, '%H:%i'),
			a.duracao_minutos,
			a.observacao,
			DATE_FORMAT(a.created_at, '%Y-%m-%d %H:%i'),

			p.id,
			p.nome,
			p.especialidade,
			p.foto_url,

			s.nome,
			s.telefone,
			s.whatsapp,
			s.endereco,
			s.cidade,
			s.estado

		FROM agendamentos a

		INNER JOIN clientes_auth_vinculos c
			ON c.auth_user_id = a.auth_user_id
		   AND c.id = ?

		INNER JOIN profissionais p
			ON p.id = a.profissional_id

		INNER JOIN saloes s
			ON s.id = a.salao_id

		WHERE a.id = ?
		  AND a.salao_id = ?

		LIMIT 1
	`

	var (
		dados         AgendamentoDados
		observacao    sql.NullString
		especialidade sql.NullString
		fotoURL       sql.NullString
		telefone      sql.NullString
		whatsapp      sql.NullString
		endereco      sql.NullString
		cidade        sql.NullString
		estado        sql.NullString
	)

	err := r.db.QueryRowContext(
		ctx,
		query,
		req.IDCliente,
		req.AgendamentoID,
		req.SalaoID,
	).Scan(
		&dados.ID,
		&dados.Status,
		&dados.Data,
		&dados.Hora,
		&dados.DuracaoMinutos,
		&observacao,
		&dados.CriadoEm,

		&dados.Profissional.ID,
		&dados.Profissional.Nome,
		&especialidade,
		&fotoURL,

		&dados.Salao.Nome,
		&telefone,
		&whatsapp,
		&endereco,
		&cidade,
		&estado,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNaoEncontrado
	}

	if err != nil {
		return nil, err
	}

	dados.Observacao = observacao.String

	dados.Profissional.Especialidade =
		especialidade.String

	dados.Profissional.FotoURL =
		fotoURL.String

	dados.Salao.Telefone =
		telefone.String

	dados.Salao.Whatsapp =
		whatsapp.String

	dados.Salao.Endereco =
		endereco.String

	dados.Salao.Cidade =
		cidade.String

	dados.Salao.Estado =
		estado.String

	return &dados, nil
}

// BuscarServicos devolve os serviços escolhidos, como foram gravados no
// momento do agendamento (nome, preço e duração do snapshot), com a
// imagem atual do serviço.
func (r *Repository) BuscarServicos(
	ctx context.Context,
	agendamentoID int64,
) ([]ServicoDados, error) {
	const query = `
		SELECT
			spa.servico_id,
			spa.nome_servico_snapshot,
			sv.img,
			spa.quantidade,
			spa.duracao_minutos,
			spa.preco_unitario
		FROM servico_por_agendamento spa
		LEFT JOIN servicos sv
			ON sv.id = spa.servico_id
		WHERE spa.agendamento_id = ?
		ORDER BY spa.id ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		agendamentoID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	servicos := make([]ServicoDados, 0)

	for rows.Next() {
		var (
			servico ServicoDados
			img     sql.NullString
		)

		if err := rows.Scan(
			&servico.ID,
			&servico.Nome,
			&img,
			&servico.Quantidade,
			&servico.DuracaoMinutos,
			&servico.PrecoUnitario,
		); err != nil {
			return nil, err
		}

		servico.FotoURL = img.String

		servicos = append(servicos, servico)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servicos, nil
}