package profissional

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

type APIResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data    any `json:"data,omitempty"`
}

func responderJSON(
	w http.ResponseWriter,
	status int,
	response APIResponse,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func lerJSON(w http.ResponseWriter, r *http.Request, destino any) bool {
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destino); err != nil {
		responderJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "JSON inválido.",
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

	err := h.service.AddRelations(r.Context(), req)

	if err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			responderJSON(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Informe salão, conta AUTH e nome válidos.",
			})

		case errors.Is(err, ErrSalaoInvalido):
			responderJSON(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Salão não encontrado.",
			})

		case errors.Is(err, ErrDuplicado):
			responderJSON(w, http.StatusConflict, APIResponse{
				Success: false,
				Message: "Esta conta já possui vínculo com o salão.",
			})

		default:
			responderJSON(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Erro ao cadastrar vínculo profissional.",
			})
		}

		return
	}

	responderJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Solicitação de vínculo enviada. Aguardando liberação.",
		Data: map[string]any{
			"id_salao":     req.IDSalao,
			"auth_user_id": req.AuthUserID,
			"status":       "aguardando_liberacao",
		},
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

	response, err := h.service.Verify(r.Context(), req)

	if err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			responderJSON(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Informe salão e conta AUTH válidos.",
			})

		default:
			responderJSON(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Erro ao verificar vínculo profissional.",
			})
		}

		return
	}

	responderJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Verificação concluída.",
		Data:    response,
	})
}