package blocks

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
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

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destino); err != nil {
		responder(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Corpo da requisição inválido.",
		})
		return false
	}

	return true
}

func (h *Handler) Listar(w http.ResponseWriter, r *http.Request) {
	var req ListRequest
	if !lerJSON(w, r, &req) {
		return
	}

	resultado, err := h.service.Listar(r.Context(), req)
	if err != nil {
		h.erro(w, err)
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    resultado,
	})
}

func (h *Handler) AdicionarDia(w http.ResponseWriter, r *http.Request) {
	var req AddDayRequest
	if !lerJSON(w, r, &req) {
		return
	}

	if err := h.service.AdicionarDia(r.Context(), req); err != nil {
		h.erro(w, err)
		return
	}

	responder(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "Dia bloqueado com sucesso!",
	})
}

func (h *Handler) AdicionarHora(w http.ResponseWriter, r *http.Request) {
	var req AddHourRequest
	if !lerJSON(w, r, &req) {
		return
	}

	if err := h.service.AdicionarHora(r.Context(), req); err != nil {
		h.erro(w, err)
		return
	}

	responder(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "Horário bloqueado com sucesso!",
	})
}

func (h *Handler) RemoverDia(w http.ResponseWriter, r *http.Request) {
	var req RemoveRequest
	if !lerJSON(w, r, &req) {
		return
	}

	if err := h.service.RemoverDia(
		r.Context(), req.IDSalao, req.ID,
	); err != nil {
		h.erro(w, err)
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Bloqueio de dia removido!",
	})
}

func (h *Handler) RemoverHora(w http.ResponseWriter, r *http.Request) {
	var req RemoveRequest
	if !lerJSON(w, r, &req) {
		return
	}

	if err := h.service.RemoverHora(
		r.Context(), req.IDSalao, req.ID,
	); err != nil {
		h.erro(w, err)
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Bloqueio de horário removido!",
	})
}

func (h *Handler) erro(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrDadosInvalidos),
		errors.Is(err, ErrDataPassada):
		responder(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": err.Error(),
		})

	case errors.Is(err, ErrProfissionalNaoEncontrado),
		errors.Is(err, ErrBloqueioNaoEncontrado):
		responder(w, http.StatusNotFound, map[string]any{
			"success": false,
			"message": err.Error(),
		})

	default:
		log.Printf("Erro no módulo de bloqueios: %v", err)

		if strings.Contains(strings.ToLower(err.Error()), "duplicate entry") {
			responder(w, http.StatusConflict, map[string]any{
				"success": false,
				"message": "Este bloqueio já existe.",
			})
			return
		}

		responder(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Erro ao processar bloqueios.",
		})
	}
}
