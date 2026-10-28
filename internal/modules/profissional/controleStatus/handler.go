package statusAG

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/profissional"
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

func lerJSON(w http.ResponseWriter, r *http.Request, destino any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(destino); err != nil {
		responder(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Corpo da requisição inválido.",
		})
		return false
	}
	return true
}

func erro(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, profissional.ErrDadosInvalidos):
		responder(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": err.Error(),
		})
	case errors.Is(err, profissional.ErrNaoEncontrado):
		responder(w, http.StatusNotFound, map[string]any{
			"success": false,
			"message": err.Error(),
		})
	default:
		log.Printf("[statusAG] erro interno: %v", err)
		responder(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Erro ao processar a requisição.",
		})
	}
}

func (h *Handler) IniciarCorte(w http.ResponseWriter, r *http.Request) {
	var req IniciarCorteRequest
	if !lerJSON(w, r, &req) {
		return
	}
	resultado, err := h.service.IniciarCorte(r.Context(), req)
	if err != nil {
		erro(w, err)
		return
	}
	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Atendimento iniciado com sucesso!",
		"data":    resultado,
	})
}

func (h *Handler) FinalizarCorte(w http.ResponseWriter, r *http.Request) {
	var req FinalizarCorteRequest
	if !lerJSON(w, r, &req) {
		return
	}
	resultado, err := h.service.FinalizarCorte(r.Context(), req)
	if err != nil {
		erro(w, err)
		return
	}
	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Atendimento concluído com sucesso!",
		"data":    resultado,
	})
}

func (h *Handler) CancelarAgendamento(w http.ResponseWriter, r *http.Request) {
	var req CancelarAgendamentoRequest
	if !lerJSON(w, r, &req) {
		return
	}
	resultado, err := h.service.CancelarAgendamento(r.Context(), req)
	if err != nil {
		erro(w, err)
		return
	}
	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Agendamento cancelado.",
		"data":    resultado,
	})
}

func (h *Handler) NaoCompareceu(w http.ResponseWriter, r *http.Request) {
	var req NaoCompareceuRequest
	if !lerJSON(w, r, &req) {
		return
	}
	resultado, err := h.service.NaoCompareceu(r.Context(), req)
	if err != nil {
		erro(w, err)
		return
	}
	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Cliente marcado como não compareceu.",
		"data":    resultado,
	})
}