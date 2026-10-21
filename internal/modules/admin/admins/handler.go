
package admins

import (
	"encoding/json"
	"errors"
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

func lerJSON(
	w http.ResponseWriter,
	r *http.Request,
	destino any,
) bool {
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

func (h *Handler) AddRelations(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req AddRelationsRequest

	if !lerJSON(w, r, &req) {
		return
	}

	if err := h.service.AddRelations(r.Context(), req); err != nil {
		h.erro(w, err)
		return
	}

	responder(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "Vínculo administrativo criado.",
	})
}

func (h *Handler) Verify(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req VerifyRequest

	if !lerJSON(w, r, &req) {
		return
	}

	resultado, err := h.service.Verify(r.Context(), req)
	if err != nil {
		h.erro(w, err)
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    resultado,
	})
}

func (h *Handler) erro(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrDadosInvalidos):
		responder(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": err.Error(),
		})

	case errors.Is(err, ErrSalaoInvalido):
		responder(w, http.StatusNotFound, map[string]any{
			"success": false,
			"message": err.Error(),
		})

	case errors.Is(err, ErrDuplicado):
		responder(w, http.StatusConflict, map[string]any{
			"success": false,
			"message": err.Error(),
		})

	default:
		responder(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Erro ao processar o vínculo administrativo.",
		})
	}
}
