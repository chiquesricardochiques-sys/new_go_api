package horarios

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

func (h *Handler) Listar(w http.ResponseWriter, r *http.Request) {
	var req ListRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responder(w, http.StatusBadRequest, false, "JSON inválido", nil)
		return
	}

	horarios, err := h.service.Listar(r.Context(), req)
	if err != nil {
		tratarErro(w, err)
		return
	}

	responder(w, http.StatusOK, true, "", map[string]any{
		"horarios": horarios,
	})
}

func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responder(w, http.StatusBadRequest, false, "JSON inválido", nil)
		return
	}

	if err := h.service.Criar(r.Context(), req); err != nil {
		tratarErro(w, err)
		return
	}

	responder(w, http.StatusOK, true, "Horário criado com sucesso!", nil)
}

func (h *Handler) Excluir(w http.ResponseWriter, r *http.Request) {
	var req DeleteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responder(w, http.StatusBadRequest, false, "JSON inválido", nil)
		return
	}

	if err := h.service.Excluir(r.Context(), req); err != nil {
		tratarErro(w, err)
		return
	}

	responder(w, http.StatusOK, true, "Horário removido com sucesso!", nil)
}

func tratarErro(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidData):
		responder(w, http.StatusBadRequest, false,
			"Informe os dados obrigatórios corretamente", nil)

	case errors.Is(err, ErrNotFound):
		responder(w, http.StatusNotFound, false,
			"Horário não encontrado", nil)

	default:
		responder(w, http.StatusInternalServerError, false,
			"Erro ao processar horários", nil)
	}
}

func responder(
	w http.ResponseWriter,
	status int,
	success bool,
	message string,
	data any,
) {
	resposta := map[string]any{
		"success": success,
	}

	if message != "" {
		resposta["message"] = message
	}

	if data != nil {
		resposta["data"] = data
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resposta)
}
