
package mysaloes

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var (
	regexUF        = regexp.MustCompile(`^[A-Z]{2}$`)
	regexNaoAlfNum = regexp.MustCompile(`[^a-z0-9]+`)

	acentos = strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c", "ñ", "n",
	)
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func truncar(s string, max int) string {
	r := []rune(strings.TrimSpace(s))

	if len(r) > max {
		r = r[:max]
	}

	return string(r)
}

func gerarSlugBase(nome string) string {
	s := acentos.Replace(strings.ToLower(strings.TrimSpace(nome)))
	s = regexNaoAlfNum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	if len(s) > 150 {
		s = strings.Trim(s[:150], "-")
	}

	return s
}

func validarAuthUser(authUserID int64) error {
	if authUserID <= 0 {
		return fmt.Errorf(
			"%w: auth_user_id obrigatório",
			ErrDadosInvalidos,
		)
	}

	return nil
}

func validarTipoPix(tipo string) error {
	if tipo == "" {
		return nil
	}

	switch tipo {
	case "cpf", "cnpj", "email", "telefone", "aleatoria":
		return nil

	default:
		return fmt.Errorf(
			"%w: tipo de chave Pix inválido",
			ErrDadosInvalidos,
		)
	}
}

// prepararDados aplica as mesmas validações e normalizações
// tanto no cadastro quanto na atualização.
func prepararDados(req CriarRequest) (*NovoSalao, string, error) {
	nome := truncar(req.Nome, 150)

	if len([]rune(nome)) < 3 {
		return nil, "", fmt.Errorf(
			"%w: informe o nome do salão (mínimo 3 caracteres)",
			ErrDadosInvalidos,
		)
	}

	estado := strings.ToUpper(strings.TrimSpace(req.Estado))

	if estado != "" && !regexUF.MatchString(estado) {
		return nil, "", fmt.Errorf(
			"%w: estado deve ter 2 letras (ex.: SP)",
			ErrDadosInvalidos,
		)
	}

	chavePixTipo := strings.ToLower(strings.TrimSpace(req.ChavePixTipo))

	if err := validarTipoPix(chavePixTipo); err != nil {
		return nil, "", err
	}

	chavePix := truncar(req.ChavePix, 255)

	if (chavePix == "") != (chavePixTipo == "") {
		return nil, "", fmt.Errorf(
			"%w: informe a chave Pix e o tipo correspondente",
			ErrDadosInvalidos,
		)
	}

	slugBase := gerarSlugBase(nome)

	if slugBase == "" {
		return nil, "", fmt.Errorf(
			"%w: o nome precisa ter letras ou números",
			ErrDadosInvalidos,
		)
	}

	dados := &NovoSalao{
		Nome:      nome,
		Descricao: truncar(req.Descricao, 65535),
		Telefone:  truncar(req.Telefone, 30),
		Email:     truncar(req.Email, 255),
		Whatsapp:  truncar(req.Whatsapp, 30),

		SiteURL:             truncar(req.SiteURL, 255),
		Instagram:           truncar(req.Instagram, 255),
		Facebook:            truncar(req.Facebook, 255),
		TipoEstabelecimento: truncar(req.TipoEstabelecimento, 100),
		Documento:           truncar(req.Documento, 20),

		ChavePix:       chavePix,
		ChavePixTipo:   chavePixTipo,
		NomeTitularPix: truncar(req.NomeTitularPix, 150),

		Endereco: truncar(req.Endereco, 255),
		Cidade:   truncar(req.Cidade, 100),
		Estado:   estado,
		CEP:      truncar(req.CEP, 12),

		// Imagens de apresentação salvas em saloes.
		LogoURL:  truncar(req.LogoURL, 500),
		ImgSalao: truncar(req.ImgSalao, 500),
	}

	return dados, slugBase, nil
}

func (s *Service) Listar(
	ctx context.Context,
	req ListarRequest,
) ([]SalaoResumo, error) {
	if err := validarAuthUser(req.AuthUserID); err != nil {
		return nil, err
	}

	return s.repository.Listar(ctx, req.AuthUserID)
}

func (s *Service) Criar(
	ctx context.Context,
	req CriarRequest,
) (*SalaoResumo, error) {
	if err := validarAuthUser(req.AuthUserID); err != nil {
		return nil, err
	}

	dados, slugBase, err := prepararDados(req)
	if err != nil {
		return nil, err
	}

	return s.repository.Criar(
		ctx,
		req.AuthUserID,
		dados,
		slugBase,
	)
}

func (s *Service) Atualizar(
	ctx context.Context,
	req AtualizarRequest,
) (*SalaoResumo, error) {
	if err := validarAuthUser(req.AuthUserID); err != nil {
		return nil, err
	}

	if req.IDSalao <= 0 {
		return nil, fmt.Errorf(
			"%w: id_salao obrigatório",
			ErrDadosInvalidos,
		)
	}

	dados, slugBase, err := prepararDados(req.CriarRequest)
	if err != nil {
		return nil, err
	}

	return s.repository.Atualizar(
		ctx,
		req.AuthUserID,
		req.IDSalao,
		dados,
		slugBase,
	)
}

// Entrar permite somente salões com status ativo.
func (s *Service) Entrar(
	ctx context.Context,
	req EntrarRequest,
) (*SalaoResumo, error) {
	if err := validarAuthUser(req.AuthUserID); err != nil {
		return nil, err
	}

	if req.IDSalao <= 0 {
		return nil, fmt.Errorf(
			"%w: id_salao obrigatório",
			ErrDadosInvalidos,
		)
	}

	salao, err := s.repository.BuscarDoAdmin(
		ctx,
		req.AuthUserID,
		req.IDSalao,
	)
	if err != nil {
		return nil, err
	}

	if salao.Status != StatusAtivo {
		return nil, ErrSalaoIndisponivel
	}

	return salao, nil
}

// AdicionarImagem valida os dados recebidos do Cloudinary
// e solicita ao repository o cadastro no banco.
func (s *Service) AdicionarImagem(
	ctx context.Context,
	req AdicionarImagemRequest,
) (*ImagemSalao, error) {
	if err := validarAuthUser(req.AuthUserID); err != nil {
		return nil, err
	}

	if req.IDSalao <= 0 {
		return nil, fmt.Errorf(
			"%w: id_salao obrigatório",
			ErrDadosInvalidos,
		)
	}

	req.URL = strings.TrimSpace(req.URL)
	req.PublicID = strings.TrimSpace(req.PublicID)
	req.Descricao = strings.TrimSpace(req.Descricao)

	if req.URL == "" || len(req.URL) > 500 {
		return nil, fmt.Errorf(
			"%w: URL da imagem inválida",
			ErrDadosInvalidos,
		)
	}

	if req.PublicID == "" || len(req.PublicID) > 255 {
		return nil, fmt.Errorf(
			"%w: public_id da imagem inválido",
			ErrDadosInvalidos,
		)
	}

	if len([]rune(req.Descricao)) > 255 {
		return nil, fmt.Errorf(
			"%w: descrição deve ter no máximo 255 caracteres",
			ErrDadosInvalidos,
		)
	}

	if req.Ordem < 0 {
		return nil, fmt.Errorf(
			"%w: ordem da imagem inválida",
			ErrDadosInvalidos,
		)
	}

	return s.repository.AdicionarImagem(ctx, req)
}

// ListarImagens retorna as imagens ativas do salão,
// desde que o administrador tenha vínculo ativo.
func (s *Service) ListarImagens(
	ctx context.Context,
	req ListarImagensRequest,
) ([]ImagemSalao, error) {
	if err := validarAuthUser(req.AuthUserID); err != nil {
		return nil, err
	}

	if req.IDSalao <= 0 {
		return nil, fmt.Errorf(
			"%w: id_salao obrigatório",
			ErrDadosInvalidos,
		)
	}

	return s.repository.ListarImagens(
		ctx,
		req.AuthUserID,
		req.IDSalao,
	)
}
// GetAtualizacao retorna todos os dados necessários para preencher
// a tela de edição do salão.
func (s *Service) GetAtualizacao(
	ctx context.Context,
	req ListarImagensRequest,
) (*DadosAtualizacaoSalao, error) {

	if err := validarAuthUser(req.AuthUserID); err != nil {
		return nil, err
	}

	if req.IDSalao <= 0 {
		return nil, fmt.Errorf(
			"%w: id_salao obrigatório",
			ErrDadosInvalidos,
		)
	}

	// Busca o salão e valida o vínculo ativo do administrador.
	salao, err := s.repository.BuscarDoAdmin(
		ctx,
		req.AuthUserID,
		req.IDSalao,
	)
	if err != nil {
		return nil, err
	}

	// Mantém a mesma regra de edição já aplicada em Atualizar.
	if salao.Status != StatusPendente &&
		salao.Status != StatusRejeitado &&
		salao.Status != StatusAtivo {
		return nil, ErrEdicaoIndisponivel
	}

	// Busca as imagens da galeria.
	imagens, err := s.repository.ListarImagens(
		ctx,
		req.AuthUserID,
		req.IDSalao,
	)
	if err != nil {
		return nil, err
	}

	return &DadosAtualizacaoSalao{
		Salao:   *salao,
		Imagens: imagens,
	}, nil
}


// AtualizarImagem valida os dados e substitui uma imagem
// já cadastrada na galeria do salão.
func (s *Service) AtualizarImagem(
	ctx context.Context,
	req AtualizarImagemRequest,
) (*ImagemSalao, error) {

	if err := validarAuthUser(req.AuthUserID); err != nil {
		return nil, err
	}

	if req.IDSalao <= 0 {
		return nil, fmt.Errorf(
			"%w: id_salao obrigatório",
			ErrDadosInvalidos,
		)
	}

	if req.IDImagem <= 0 {
		return nil, fmt.Errorf(
			"%w: id_imagem obrigatório",
			ErrDadosInvalidos,
		)
	}

	req.URL = strings.TrimSpace(req.URL)
	req.PublicID = strings.TrimSpace(req.PublicID)

	if req.URL == "" || len(req.URL) > 500 {
		return nil, fmt.Errorf(
			"%w: URL da imagem inválida",
			ErrDadosInvalidos,
		)
	}

	if req.PublicID == "" || len(req.PublicID) > 255 {
		return nil, fmt.Errorf(
			"%w: public_id da imagem inválido",
			ErrDadosInvalidos,
		)
	}

	return s.repository.AtualizarImagem(ctx, req)
}