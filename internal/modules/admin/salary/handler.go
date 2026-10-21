package salary

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

func (h *Handler) Rendimento(w http.ResponseWriter, r *http.Request) {
	var req Request

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeResponse(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "JSON inválido.",
		})
		return
	}

	data, err := h.service.Rendimento(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidData):
			writeResponse(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"message": "Informe um salão, profissional e mês válidos.",
			})
		case errors.Is(err, ErrNotFound):
			writeResponse(w, http.StatusNotFound, map[string]any{
				"success": false,
				"message": "Profissional não encontrado.",
			})
		default:
			writeResponse(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"message": "Erro ao buscar rendimento do profissional.",
			})
		}
		return
	}

	writeResponse(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
	})
}

func writeResponse(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
