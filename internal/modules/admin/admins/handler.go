package admins

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

// POST /api/admin/admins/verify
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

	default:
		log.Println("erro ao verificar vínculo administrativo:", err)

		responder(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Erro ao processar o vínculo administrativo.",
		})
	}
}