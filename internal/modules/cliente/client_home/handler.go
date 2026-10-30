package client_home

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

	var req ClientHomeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		_ = json.NewEncoder(w).Encode(
			map[string]any{
				"success": false,
				"message": "JSON inválido.",
			},
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
					"message": "salao_id e id_cliente são obrigatórios.",
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

		case errors.Is(err, ErrClienteNaoEncontrado):
			w.WriteHeader(http.StatusNotFound)

			_ = json.NewEncoder(w).Encode(
				map[string]any{
					"success": false,
					"message": "Cliente não encontrado.",
				},
			)

		default:
			w.WriteHeader(http.StatusInternalServerError)

			_ = json.NewEncoder(w).Encode(
				map[string]any{
					"success": false,
					"message": "Erro interno ao carregar a Home do cliente.",
				},
			)
		}

		return
	}

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(
		ClientHomeResponse{
			Success: true,
			Data:    data,
		},
	)
}