package textos_salao

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(
	service *Service,
) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req ListRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"dados inválidos.",
		)
		return
	}

	textos, err := h.service.Listar(
		r.Context(),
		&req,
	)

	if err != nil {
		handleServiceError(
			w,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"success": true,
			"data": map[string]any{
				"textos": textos,
			},
		},
	)
}

func (h *Handler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req GetRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"dados inválidos.",
		)
		return
	}

	texto, err := h.service.Buscar(
		r.Context(),
		&req,
	)

	if err != nil {
		handleServiceError(
			w,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"success": true,
			"data": map[string]any{
				"texto": texto,
			},
		},
	)
}

func (h *Handler) Add(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req AddRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"dados inválidos.",
		)
		return
	}

	texto, err := h.service.Adicionar(
		r.Context(),
		&req,
	)

	if err != nil {
		handleServiceError(
			w,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		map[string]any{
			"success": true,
			"message": "Texto adicionado com sucesso.",
			"data": map[string]any{
				"texto": texto,
			},
		},
	)
}

func (h *Handler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req UpdateRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"dados inválidos.",
		)
		return
	}

	texto, err := h.service.Atualizar(
		r.Context(),
		&req,
	)

	if err != nil {
		handleServiceError(
			w,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"success": true,
			"message": "Texto atualizado com sucesso.",
			"data": map[string]any{
				"texto": texto,
			},
		},
	)
}

func (h *Handler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req DeleteRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"dados inválidos.",
		)
		return
	}

	err := h.service.Excluir(
		r.Context(),
		&req,
	)

	if err != nil {
		handleServiceError(
			w,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"success": true,
			"message": "Texto excluído com sucesso.",
		},
	)
}

func handleServiceError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(
		err,
		ErrInvalidData,
	):
		writeError(
			w,
			http.StatusBadRequest,
			"dados inválidos.",
		)

	case errors.Is(
		err,
		ErrNotFound,
	):
		writeError(
			w,
			http.StatusNotFound,
			"Texto do salão não encontrado.",
		)

	case errors.Is(
		err,
		ErrDuplicate,
	):
		writeError(
			w,
			http.StatusConflict,
			"Já existe um texto com essa chave para este salão.",
		)

	default:
		writeError(
			w,
			http.StatusInternalServerError,
			"Erro interno do servidor.",
		)
	}
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(
		w,
	).Encode(data)
}

func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(
		w,
		status,
		map[string]any{
			"success": false,
			"message": message,
		},
	)
}