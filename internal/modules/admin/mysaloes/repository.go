
package mysaloes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
)

const (
	limitePendentesPorConta = 3
	maxTentativasSlug       = 20
	limiteImagensSalao      = 3
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Salões vinculados ao administrador.
const selecaoSalao = `
	SELECT
		s.id,
		s.nome,
		s.descricao,
		s.slug,
		s.codigo_painel,
		s.telefone,
		s.email,
		s.whatsapp,
		s.site_url,
		s.instagram,
		s.facebook,
		s.tipo_estabelecimento,
		s.documento,
		s.chave_pix,
		s.chave_pix_tipo,
		s.nome_titular_pix,
		s.endereco,
		s.cidade,
		s.estado,
		s.cep,
		s.latitude,
		s.longitude,
		s.logo_url,
		s.img_salao,
		s.status,
		s.motivo_analise,
		sa.papel,

		CASE
			WHEN EXISTS (
				SELECT 1
				FROM avisos_salao av
				WHERE av.salao_id = s.id
				  AND av.lido = 1
			)
			THEN 1
			ELSE 0
		END AS tem_avisos_nao_lidos,

		s.img_salao AS imagem_card

	FROM salao_admins sa
	JOIN saloes s ON s.id = sa.salao_id

	WHERE sa.auth_user_id = ?
	  AND sa.status = 'ativo'
`

type scanner interface {
	Scan(dest ...any) error
}

func escanear(row scanner) (SalaoResumo, error) {
	var s SalaoResumo

	err := row.Scan(
		&s.ID,
		&s.Nome,
		&s.Descricao,
		&s.Slug,
		&s.CodigoPainel,
		&s.Telefone,
		&s.Email,
		&s.Whatsapp,
		&s.SiteURL,
		&s.Instagram,
		&s.Facebook,
		&s.TipoEstabelecimento,
		&s.Documento,
		&s.ChavePix,
		&s.ChavePixTipo,
		&s.NomeTitularPix,
		&s.Endereco,
		&s.Cidade,
		&s.Estado,
		&s.CEP,
		&s.Latitude,
		&s.Longitude,
		&s.LogoURL,
		&s.ImgSalao,
		&s.Status,
		&s.MotivoAnalise,
		&s.Papel,
		&s.TemAvisosNaoLidos,
		&s.ImagemCard,
	)

	return s, err
}

func nulo(s string) any {
	s = strings.TrimSpace(s)

	if s == "" {
		return nil
	}

	return s
}

func slugDuplicado(err error) bool {
	var me *mysql.MySQLError

	return errors.As(err, &me) &&
		me.Number == 1062 &&
		strings.Contains(me.Message, "uk_saloes_slug")
}

// Listar retorna os salões vinculados ao administrador.
func (r *Repository) Listar(
	ctx context.Context,
	authUserID int64,
) ([]SalaoResumo, error) {
	rows, err := r.db.QueryContext(
		ctx,
		selecaoSalao+" ORDER BY s.nome ASC",
		authUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lista := make([]SalaoResumo, 0)

	for rows.Next() {
		s, err := escanear(rows)
		if err != nil {
			return nil, err
		}

		lista = append(lista, s)
	}

	return lista, rows.Err()
}

// BuscarDoAdmin devolve o salão somente se o administrador
// possui vínculo ativo com ele.
func (r *Repository) BuscarDoAdmin(
	ctx context.Context,
	authUserID int64,
	idSalao int64,
) (*SalaoResumo, error) {
	s, err := escanear(r.db.QueryRowContext(
		ctx,
		selecaoSalao+" AND s.id = ? LIMIT 1",
		authUserID,
		idSalao,
	))

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSemAcesso
	}

	if err != nil {
		return nil, err
	}

	return &s, nil
}

// Criar cria o salão pendente e o vínculo dono na mesma transação.
func (r *Repository) Criar(
	ctx context.Context,
	authUserID int64,
	d *NovoSalao,
	slugBase string,
) (*SalaoResumo, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var pendentes int

	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM salao_admins sa
		JOIN saloes s ON s.id = sa.salao_id
		WHERE sa.auth_user_id = ?
		  AND sa.papel = 'dono'
		  AND s.status = 'pendente'
	`, authUserID).Scan(&pendentes)

	if err != nil {
		return nil, err
	}

	if pendentes >= limitePendentesPorConta {
		return nil, ErrLimitePendentes
	}

	var idSalao int64
	slug := slugBase

	for tentativa := 1; tentativa <= maxTentativasSlug; tentativa++ {
		if tentativa > 1 {
			slug = fmt.Sprintf("%s-%d", slugBase, tentativa)
		}

		res, err := tx.ExecContext(ctx, `
			INSERT INTO saloes (
				nome,
				descricao,
				slug,
				telefone,
				email,
				whatsapp,
				site_url,
				instagram,
				facebook,
				tipo_estabelecimento,
				documento,
				chave_pix,
				chave_pix_tipo,
				nome_titular_pix,
				endereco,
				cidade,
				estado,
				cep,
				logo_url,
				img_salao,
				status
			)
			VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 
				'pendente'
			)
		`,
			d.Nome,
			nulo(d.Descricao),
			slug,
			nulo(d.Telefone),
			nulo(d.Email),
			nulo(d.Whatsapp),
			nulo(d.SiteURL),
			nulo(d.Instagram),
			nulo(d.Facebook),
			nulo(d.TipoEstabelecimento),
			nulo(d.Documento),
			nulo(d.ChavePix),
			nulo(d.ChavePixTipo),
			nulo(d.NomeTitularPix),
			nulo(d.Endereco),
			nulo(d.Cidade),
			nulo(d.Estado),
			nulo(d.CEP),
			nulo(d.LogoURL),
			nulo(d.ImgSalao),
		)

		if err != nil {
			if slugDuplicado(err) {
				continue
			}

			return nil, err
		}

		idSalao, err = res.LastInsertId()
		if err != nil {
			return nil, err
		}

		break
	}

	if idSalao == 0 {
		return nil, ErrSlugIndisponivel
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO salao_admins (
			salao_id,
			auth_user_id,
			papel,
			status
		)
		VALUES (?, ?, 'dono', 'ativo')
	`, idSalao, authUserID)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.BuscarDoAdmin(ctx, authUserID, idSalao)
}

// Atualizar altera os dados do salão existente.
// Somente salões pendentes ou rejeitados podem ser reenviados para análise.
func (r *Repository) Atualizar(
	ctx context.Context,
	authUserID int64,
	idSalao int64,
	d *NovoSalao,
	slugBase string,
) (*SalaoResumo, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var status string

	err = tx.QueryRowContext(ctx, `
		SELECT s.status
		FROM salao_admins sa
		JOIN saloes s ON s.id = sa.salao_id
		WHERE sa.auth_user_id = ?
		  AND sa.salao_id = ?
		  AND sa.status = 'ativo'
		LIMIT 1
		FOR UPDATE
	`, authUserID, idSalao).Scan(&status)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSemAcesso
	}

	if err != nil {
		return nil, err
	}

	if status != StatusPendente &&
		status != StatusRejeitado &&
		status != StatusAtivo {
		return nil, ErrEdicaoIndisponivel
	}

	var atualizado bool

	for tentativa := 1; tentativa <= maxTentativasSlug; tentativa++ {
		slug := slugBase

		if tentativa > 1 {
			slug = fmt.Sprintf("%s-%d", slugBase, tentativa)
		}

		_, err := tx.ExecContext(ctx, `
			UPDATE saloes
			SET
				nome = ?,
				descricao = ?,
				slug = ?,
				telefone = ?,
				email = ?,
				whatsapp = ?,
				site_url = ?,
				instagram = ?,
				facebook = ?,
				tipo_estabelecimento = ?,
				documento = ?,
				chave_pix = ?,
				chave_pix_tipo = ?,
				nome_titular_pix = ?,
				endereco = ?,
				cidade = ?,
				estado = ?,
				cep = ?,
				logo_url = ?,
				img_salao = ?,
				status = CASE
					WHEN ? = 'ativo' THEN 'ativo'
					ELSE 'pendente'
				END,
				motivo_analise = CASE
					WHEN ? = 'ativo' THEN motivo_analise
					ELSE NULL
				END
				WHERE id = ?
		`,
			d.Nome,
			nulo(d.Descricao),
			slug,
			nulo(d.Telefone),
			nulo(d.Email),
			nulo(d.Whatsapp),
			nulo(d.SiteURL),
			nulo(d.Instagram),
			nulo(d.Facebook),
			nulo(d.TipoEstabelecimento),
			nulo(d.Documento),
			nulo(d.ChavePix),
			nulo(d.ChavePixTipo),
			nulo(d.NomeTitularPix),
			nulo(d.Endereco),
			nulo(d.Cidade),
			nulo(d.Estado),
			nulo(d.CEP),
			nulo(d.LogoURL),
			nulo(d.ImgSalao),
			status,
			status,
			idSalao,
		)

		if err != nil {
			if slugDuplicado(err) {
				continue
			}

			return nil, err
		}

		atualizado = true
		break
	}

	if !atualizado {
		return nil, ErrSlugIndisponivel
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.BuscarDoAdmin(ctx, authUserID, idSalao)
}

// AdicionarImagem salva a URL e o public_id de uma imagem
// que já foi enviada ao Cloudinary.
func (r *Repository) AdicionarImagem(
	ctx context.Context,
	req AdicionarImagemRequest,
) (*ImagemSalao, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Bloqueia a linha do salão para impedir que uploads simultâneos
	// ultrapassem o limite de três imagens.
	var idSalao int64

	err = tx.QueryRowContext(ctx, `
		SELECT s.id
		FROM saloes s
		JOIN salao_admins sa
			ON sa.salao_id = s.id
		WHERE s.id = ?
		  AND sa.auth_user_id = ?
		  AND sa.status = 'ativo'
		LIMIT 1
		FOR UPDATE
	`,
		req.IDSalao,
		req.AuthUserID,
	).Scan(&idSalao)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSemAcesso
	}

	if err != nil {
		return nil, err
	}

	var total int

	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM imagens_salao
		WHERE salao_id = ?
		  AND ativo = 1
	`, idSalao).Scan(&total)

	if err != nil {
		return nil, err
	}

	if total >= limiteImagensSalao {
		return nil, ErrLimiteImagens
	}

	// A primeira imagem é principal automaticamente.
	principal := req.Principal || total == 0

	if principal {
		_, err = tx.ExecContext(ctx, `
			UPDATE imagens_salao
			SET principal = 0
			WHERE salao_id = ?
			  AND ativo = 1
		`, idSalao)

		if err != nil {
			return nil, err
		}
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO imagens_salao (
			salao_id,
			url,
			public_id,
			descricao,
			ordem,
			principal,
			ativo
		)
		VALUES (?, ?, ?, ?, ?, ?, 1)
	`,
		idSalao,
		req.URL,
		req.PublicID,
		nulo(req.Descricao),
		req.Ordem,
		principal,
	)

	if err != nil {
		return nil, err
	}

	imagemID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	var imagem ImagemSalao

	err = tx.QueryRowContext(ctx, `
		SELECT
			id,
			salao_id,
			url,
			public_id,
			descricao,
			ordem,
			principal
		FROM imagens_salao
		WHERE id = ?
		  AND salao_id = ?
	`,
		imagemID,
		idSalao,
	).Scan(
		&imagem.ID,
		&imagem.IDSalao,
		&imagem.URL,
		&imagem.PublicID,
		&imagem.Descricao,
		&imagem.Ordem,
		&imagem.Principal,
	)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &imagem, nil
}

// ListarImagens retorna as imagens ativas do salão.
// Antes da consulta, verifica se o administrador possui vínculo ativo.
func (r *Repository) ListarImagens(
	ctx context.Context,
	authUserID int64,
	idSalao int64,
) ([]ImagemSalao, error) {
	var existe int

	err := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM salao_admins
		WHERE auth_user_id = ?
		  AND salao_id = ?
		  AND status = 'ativo'
		LIMIT 1
	`, authUserID, idSalao).Scan(&existe)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSemAcesso
	}

	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			salao_id,
			url,
			public_id,
			descricao,
			ordem,
			principal
		FROM imagens_salao
		WHERE salao_id = ?
		  AND ativo = 1
		ORDER BY
			principal DESC,
			ordem ASC,
			id ASC
		LIMIT ?
	`, idSalao, limiteImagensSalao)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	imagens := make([]ImagemSalao, 0)

	for rows.Next() {
		var imagem ImagemSalao

		err := rows.Scan(
			&imagem.ID,
			&imagem.IDSalao,
			&imagem.URL,
			&imagem.PublicID,
			&imagem.Descricao,
			&imagem.Ordem,
			&imagem.Principal,
		)

		if err != nil {
			return nil, err
		}

		imagens = append(imagens, imagem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return imagens, nil
}

// AtualizarImagem substitui a URL e o public_id de uma imagem existente.
// Não cria um novo registro na galeria.
func (r *Repository) AtualizarImagem(
	ctx context.Context,
	req AtualizarImagemRequest,
) (*ImagemSalao, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Confirma o vínculo ativo do administrador e bloqueia o salão.
	var idSalao int64

	err = tx.QueryRowContext(ctx, `
		SELECT s.id
		FROM saloes s
		JOIN salao_admins sa
			ON sa.salao_id = s.id
		WHERE s.id = ?
		  AND sa.auth_user_id = ?
		  AND sa.status = 'ativo'
		LIMIT 1
		FOR UPDATE
	`,
		req.IDSalao,
		req.AuthUserID,
	).Scan(&idSalao)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSemAcesso
	}

	if err != nil {
		return nil, err
	}

	// Atualiza somente a imagem pertencente ao salão informado.
	_, err = tx.ExecContext(ctx, `
		UPDATE imagens_salao
		SET
			url = ?,
			public_id = ?
		WHERE id = ?
		  AND salao_id = ?
		  AND ativo = 1
	`,
		req.URL,
		req.PublicID,
		req.IDImagem,
		idSalao,
	)

	if err != nil {
		return nil, err
	}

	// Recupera a imagem atualizada.
	var imagem ImagemSalao

	err = tx.QueryRowContext(ctx, `
		SELECT
			id,
			salao_id,
			url,
			public_id,
			descricao,
			ordem,
			principal
		FROM imagens_salao
		WHERE id = ?
		  AND salao_id = ?
		  AND ativo = 1
		LIMIT 1
	`,
		req.IDImagem,
		idSalao,
	).Scan(
		&imagem.ID,
		&imagem.IDSalao,
		&imagem.URL,
		&imagem.PublicID,
		&imagem.Descricao,
		&imagem.Ordem,
		&imagem.Principal,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrImagemNaoEncontrada
	}

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &imagem, nil
}