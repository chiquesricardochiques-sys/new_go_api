package home

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

func (h *Handler) Home(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	var req HomeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			`{"success":false,"message":"JSON inválido."}`,
			http.StatusBadRequest,
		)
		return
	}

	data, err := h.service.Buscar(
		r.Context(),
		req,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			w.WriteHeader(http.StatusBadRequest)

			_ = json.NewEncoder(w).Encode(
				map[string]any{
					"success": false,
					"message": "salao_id é obrigatório.",
				},
			)

		case errors.Is(err, ErrSalaoNaoEncontrado):
			w.WriteHeader(http.StatusNotFound)

			_ = json.NewEncoder(w).Encode(
				map[string]any{
					"success": false,
					"message": "Salão não encontrado ou inativo.",
				},
			)

		default:
			w.WriteHeader(http.StatusInternalServerError)

			_ = json.NewEncoder(w).Encode(
				map[string]any{
					"success": false,
					"message": "Erro interno ao carregar a Home.",
				},
			)
		}

		return
	}

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(
		HomeResponse{
			Success: true,
			Data:    data,
		},
	)
}