package mysaloes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func responder(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

func falha(w http.ResponseWriter, status int, mensagem string) {
	responder(w, status, map[string]any{
		"success": false,
		"message": mensagem,
	})
}

func lerJSON(w http.ResponseWriter, r *http.Request, destino any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destino); err != nil {
		falha(w, http.StatusBadRequest, "Corpo da requisição inválido.")
		return false
	}

	return true
}

func (h *Handler) erro(w http.ResponseWriter, err error, contexto string) {
	switch {
	case errors.Is(err, ErrDadosInvalidos):
		falha(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, ErrSemAcesso):
		falha(w, http.StatusNotFound, err.Error())

	case errors.Is(err, ErrSalaoIndisponivel):
		falha(w, http.StatusForbidden, err.Error())

	case errors.Is(err, ErrEdicaoIndisponivel),
		errors.Is(err, ErrLimitePendentes),
		errors.Is(err, ErrSlugIndisponivel),
		errors.Is(err, ErrLimiteImagens):

		falha(w, http.StatusConflict, err.Error())

	default:
		log.Println(contexto, err)

		falha(
			w,
			http.StatusInternalServerError,
			"Erro interno ao processar a solicitação.",
		)
	}
}

// POST /api/admin/meus-saloes/listar
func (h *Handler) Listar(w http.ResponseWriter, r *http.Request) {
	var req ListarRequest

	if !lerJSON(w, r, &req) {
		return
	}

	saloes, err := h.service.Listar(r.Context(), req)
	if err != nil {
		h.erro(w, err, "erro ao listar salões do admin:")
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"saloes": saloes,
		},
	})
}

// POST /api/admin/meus-saloes/criar
func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	var req CriarRequest

	if !lerJSON(w, r, &req) {
		return
	}

	salao, err := h.service.Criar(r.Context(), req)
	if err != nil {
		h.erro(w, err, "erro ao criar salão:")
		return
	}

	responder(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "Salão criado. Ele ficará aguardando liberação.",
		"data": map[string]any{
			"salao": salao,
		},
	})
}

// POST /api/admin/meus-saloes/atualizar
func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	var req AtualizarRequest

	if !lerJSON(w, r, &req) {
		return
	}

	salao, err := h.service.Atualizar(r.Context(), req)
	if err != nil {
		h.erro(w, err, "erro ao atualizar salão:")
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Dados atualizados. O salão foi enviado novamente para análise.",
		"data": map[string]any{
			"salao": salao,
		},
	})
}

// POST /api/admin/meus-saloes/entrar
func (h *Handler) Entrar(w http.ResponseWriter, r *http.Request) {
	var req EntrarRequest

	if !lerJSON(w, r, &req) {
		return
	}

	salao, err := h.service.Entrar(r.Context(), req)
	if err != nil {
		h.erro(w, err, "erro ao entrar no salão:")
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"salao": salao,
		},
	})
}

// POST /api/admin/meus-saloes/imagens/adicionar
func (h *Handler) AdicionarImagem(w http.ResponseWriter, r *http.Request) {
	var req AdicionarImagemRequest

	if !lerJSON(w, r, &req) {
		return
	}

	imagem, err := h.service.AdicionarImagem(r.Context(), req)
	if err != nil {
		h.erro(w, err, "erro ao adicionar imagem do salão:")
		return
	}

	responder(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "Imagem adicionada com sucesso.",
		"data": map[string]any{
			"imagem": imagem,
		},
	})
}

// POST /api/admin/meus-saloes/imagens/listar
func (h *Handler) ListarImagens(w http.ResponseWriter, r *http.Request) {
	var req ListarImagensRequest

	if !lerJSON(w, r, &req) {
		return
	}

	imagens, err := h.service.ListarImagens(r.Context(), req)
	if err != nil {
		h.erro(w, err, "erro ao listar imagens do salão:")
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"imagens": imagens,
		},
	})
}
// POST /api/admin/meus-saloes/atualizacao
func (h *Handler) GetAtualizacao(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req ListarImagensRequest

	if !lerJSON(w, r, &req) {
		return
	}

	dados, err := h.service.GetAtualizacao(r.Context(), req)
	if err != nil {
		h.erro(w, err, "erro ao buscar dados para atualização do salão:")
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    dados,
	})
}


// POST /api/admin/meus-saloes/imagens/atualizar
func (h *Handler) AtualizarImagem(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req AtualizarImagemRequest

	if !lerJSON(w, r, &req) {
		return
	}

	imagem, err := h.service.AtualizarImagem(r.Context(), req)
	if err != nil {
		h.erro(w, err, "erro ao atualizar imagem do salão:")
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Imagem atualizada com sucesso.",
		"data": map[string]any{
			"imagem": imagem,
		},
	})
}