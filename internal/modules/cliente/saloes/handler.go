package saloes

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

func escreverJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func falha(w http.ResponseWriter, status int, mensagem string) {
	escreverJSON(w, status, map[string]any{
		"success": false,
		"message": mensagem,
	})
}

func (h *Handler) responderErro(w http.ResponseWriter, err error, contexto string) {
	switch {
	case errors.Is(err, ErrDadosInvalidos):
		falha(w, http.StatusBadRequest, "Dados inválidos.")
	case errors.Is(err, ErrSalaoNaoEncontrado):
		falha(w, http.StatusNotFound, "Salão não encontrado ou inativo.")
	case errors.Is(err, ErrNaoPodeAvaliar):
		falha(w, http.StatusConflict, "Este atendimento não pode ser avaliado.")
	default:
		log.Println(contexto, err)
		falha(w, http.StatusInternalServerError, "Erro interno ao processar a solicitação.")
	}
}

func lerJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
		return false
	}

	return true
}

// POST /saloes/buscar
func (h *Handler) Buscar(w http.ResponseWriter, r *http.Request) {
	var req BuscarRequest
	if !lerJSON(w, r, &req) {
		return
	}

	data, err := h.service.Buscar(r.Context(), req)
	if err != nil {
		h.responderErro(w, err, "erro ao buscar salões:")
		return
	}

	escreverJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
	})
}

// POST /saloes/por-slug
func (h *Handler) PorSlug(w http.ResponseWriter, r *http.Request) {
	var req PorSlugRequest
	if !lerJSON(w, r, &req) {
		return
	}

	salao, err := h.service.PorSlug(r.Context(), req)
	if err != nil {
		h.responderErro(w, err, "erro ao buscar salão por slug:")
		return
	}

	escreverJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"salao": salao,
		},
	})
}

// POST /saloes/detalhes
func (h *Handler) Detalhes(w http.ResponseWriter, r *http.Request) {
	var req PorSlugRequest
	if !lerJSON(w, r, &req) {
		return
	}

	data, err := h.service.Detalhes(r.Context(), req)
	if err != nil {
		h.responderErro(w, err, "erro ao buscar detalhes do salão:")
		return
	}

	escreverJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
	})
}

// POST /saloes/avaliar
func (h *Handler) Avaliar(w http.ResponseWriter, r *http.Request) {
	var req AvaliarRequest
	if !lerJSON(w, r, &req) {
		return
	}

	if err := h.service.Avaliar(r.Context(), req); err != nil {
		h.responderErro(w, err, "erro ao avaliar atendimento:")
		return
	}

	escreverJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Avaliação registrada.",
	})
}

// POST /saloes/cidades
func (h *Handler) Cidades(w http.ResponseWriter, r *http.Request) {
	cidades, err := h.service.Cidades(r.Context())
	if err != nil {
		h.responderErro(w, err, "erro ao listar cidades:")
		return
	}

	escreverJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"cidades": cidades,
		},
	})
}