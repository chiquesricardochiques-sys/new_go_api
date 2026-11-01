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
	default:
		log.Println(contexto, err)
		falha(w, http.StatusInternalServerError, "Erro interno ao processar a solicitação.")
	}
}

// POST /saloes/buscar
func (h *Handler) Buscar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req BuscarRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
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
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req PorSlugRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
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