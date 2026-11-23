package codeEnt

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) BuscarSalao(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req BuscarSalaoRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Requisição inválida.",
		})
		return
	}

	salao, err := h.service.BuscarSalao(
		r.Context(),
		req.CodigoPainel,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrCodigoObrigatorio):
			w.WriteHeader(http.StatusBadRequest)

		case errors.Is(err, ErrSalaoNaoEncontrado):
			w.WriteHeader(http.StatusNotFound)

		case errors.Is(err, ErrSalaoIndisponivel):
			w.WriteHeader(http.StatusForbidden)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(BuscarSalaoResponse{
		Success: true,
		Message: "Salão localizado com sucesso.",
		Salao:   salao,
	})
}