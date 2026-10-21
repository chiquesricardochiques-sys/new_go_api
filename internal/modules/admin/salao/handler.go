package salao

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

func decodeJSON(
	r *http.Request,
	dst any,
) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(dst)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func (h *Handler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req GetRequest

	if err := decodeJSON(r, &req); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"success": false,
				"message": "JSON inválido.",
			},
		)

		return
	}

	salao, err := h.service.Buscar(
		r.Context(),
		&req,
	)

	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"success": true,
			"data": map[string]any{
				"salao": salao,
			},
		},
	)
}

func (h *Handler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req UpdateRequest

	if err := decodeJSON(r, &req); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"success": false,
				"message": "JSON inválido.",
			},
		)

		return
	}

	if err := h.service.Atualizar(
		r.Context(),
		&req,
	); err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"success": true,
			"message": "Salão atualizado com sucesso.",
		},
	)
}

func (h *Handler) handleError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrInvalidData):
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"success": false,
				"message": "Dados inválidos.",
			},
		)

	case errors.Is(err, ErrNotFound):
		writeJSON(
			w,
			http.StatusNotFound,
			map[string]any{
				"success": false,
				"message": "Salão não encontrado.",
			},
		)

	default:
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]any{
				"success": false,
				"message": "Erro interno ao processar a solicitação.",
			},
		)
	}
}