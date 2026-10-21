package cliente

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
	return &Handler{
		service: service,
	}
}

type response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func responderJSON(
	w http.ResponseWriter,
	status int,
	resp response,
) {
	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(resp)
}

func lerRequest(
	w http.ResponseWriter,
	r *http.Request,
) (RelacaoRequest, bool) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1<<20,
	)

	var req RelacaoRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		responderJSON(
			w,
			http.StatusBadRequest,
			response{
				Success: false,
				Message: "Corpo da requisição inválido.",
			},
		)

		return RelacaoRequest{}, false
	}

	return req, true
}

func (h *Handler) Verificar(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		w.Header().Set(
			"Allow",
			http.MethodPost,
		)

		responderJSON(
			w,
			http.StatusMethodNotAllowed,
			response{
				Success: false,
				Message: "Método não permitido.",
			},
		)

		return
	}

	req, ok := lerRequest(w, r)
	if !ok {
		return
	}

	relacao, err := h.service.Verificar(
		r.Context(),
		req,
	)

	if err != nil {
		h.responderErro(w, err)
		return
	}

	responderJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Data: relacao,
		},
	)
}

func (h *Handler) Adicionar(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		w.Header().Set(
			"Allow",
			http.MethodPost,
		)

		responderJSON(
			w,
			http.StatusMethodNotAllowed,
			response{
				Success: false,
				Message: "Método não permitido.",
			},
		)

		return
	}

	req, ok := lerRequest(w, r)
	if !ok {
		return
	}

	relacao, criado, err := h.service.Adicionar(
		r.Context(),
		req,
	)

	if err != nil {
		h.responderErro(w, err)
		return
	}

	message := "Relação do cliente já existia."

	if criado {
		message = "Relação do cliente criada com sucesso."
	}

	responderJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Message: message,
			Data:    relacao,
		},
	)
}

func (h *Handler) responderErro(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrDadosInvalidos):
		responderJSON(
			w,
			http.StatusBadRequest,
			response{
				Success: false,
				Message: err.Error(),
			},
		)

	case errors.Is(err, ErrSalaoNaoEncontrado):
		responderJSON(
			w,
			http.StatusNotFound,
			response{
				Success: false,
				Message: "Salão não encontrado ou inativo.",
			},
		)

	case errors.Is(err, ErrRelacaoNaoEncontrada):
		responderJSON(
			w,
			http.StatusNotFound,
			response{
				Success: false,
				Message: "Cliente ainda não possui relação interna.",
			},
		)

	default:
		log.Printf(
			"Erro na relação do cliente: %v",
			err,
		)

		responderJSON(
			w,
			http.StatusInternalServerError,
			response{
				Success: false,
				Message: "Erro interno ao processar a relação do cliente.",
			},
		)
	}
}