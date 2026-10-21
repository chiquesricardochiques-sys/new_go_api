package auth

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

func responderJSON(
	w http.ResponseWriter,
	status int,
	resposta LoginResponse,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resposta)
}

func lerLogin(w http.ResponseWriter, r *http.Request) (LoginRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req LoginRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		responderJSON(w, http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "Corpo da requisição inválido.",
		})
		return LoginRequest{}, false
	}

	return req, true
}

func (h *Handler) LoginAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		responderJSON(w, http.StatusMethodNotAllowed, LoginResponse{
			Success: false,
			Message: "Método não permitido.",
		})
		return
	}

	req, ok := lerLogin(w, r)
	if !ok {
		return
	}

	admin, err := h.service.LoginAdmin(r.Context(), req)
	if err != nil {
		h.responderErro(w, err)
		return
	}

	responderJSON(w, http.StatusOK, LoginResponse{
		Success: true,
		Data: map[string]any{
			"admin": admin,
		},
	})
}

func (h *Handler) LoginProfissional(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		responderJSON(w, http.StatusMethodNotAllowed, LoginResponse{
			Success: false,
			Message: "Método não permitido.",
		})
		return
	}

	req, ok := lerLogin(w, r)
	if !ok {
		return
	}

	profissional, err := h.service.LoginProfissional(r.Context(), req)
	if err != nil {
		h.responderErro(w, err)
		return
	}

	responderJSON(w, http.StatusOK, LoginResponse{
		Success: true,
		Data: map[string]any{
			"profissional": profissional,
		},
	})
}

func (h *Handler) responderErro(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrDadosInvalidos),
		errors.Is(err, ErrSalaoInvalido):
		responderJSON(w, http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: err.Error(),
		})

	case errors.Is(err, ErrCredenciais):
		responderJSON(w, http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "Email ou senha incorretos.",
		})

	case errors.Is(err, ErrProfInativo):
		responderJSON(w, http.StatusForbidden, LoginResponse{
			Success: false,
			Message: err.Error(),
		})

	default:
		log.Printf("Erro interno no login: %v", err)
		responderJSON(w, http.StatusInternalServerError, LoginResponse{
			Success: false,
			Message: "Erro interno ao realizar login.",
		})
	}
}
