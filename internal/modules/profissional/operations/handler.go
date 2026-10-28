package profissional

import (
	"context"
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

func responder(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// lerJSON NÃO usa DisallowUnknownFields de propósito: durante a migração o app
// ainda pode mandar campos antigos (project_id, instance_id...) e eles são ignorados.
func lerJSON(w http.ResponseWriter, r *http.Request, destino any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if err := json.NewDecoder(r.Body).Decode(destino); err != nil {
		responder(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Corpo da requisição inválido.",
		})
		return false
	}

	return true
}

// comData monta um handler padrão: lê o JSON, chama o service e responde
// {"success": true, "data": <resultado>}.
func comData[Req any, Res any](fn func(ctx context.Context, req Req) (Res, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Req

		if !lerJSON(w, r, &req) {
			return
		}

		resultado, err := fn(r.Context(), req)
		if err != nil {
			erro(w, err)
			return
		}

		responder(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    resultado,
		})
	}
}

// AlterarStatus é separado porque o Node devolvia também "message".
func (h *Handler) AlterarStatus(w http.ResponseWriter, r *http.Request) {
	var req AlterarStatusRequest

	if !lerJSON(w, r, &req) {
		return
	}

	resultado, err := h.service.AlterarStatus(r.Context(), req)
	if err != nil {
		erro(w, err)
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Status atualizado com sucesso!",
		"data":    resultado,
	})
}

func erro(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrDadosInvalidos):
		responder(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": err.Error(),
		})

	case errors.Is(err, ErrNaoEncontrado):
		responder(w, http.StatusNotFound, map[string]any{
			"success": false,
			"message": err.Error(),
		})

	case errors.Is(err, ErrProfissionalInativo):
		responder(w, http.StatusForbidden, map[string]any{
			"success": false,
			"message": err.Error(),
		})

	default:
		log.Printf("[profissional] erro interno: %v", err)
		responder(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Erro ao processar a requisição do profissional.",
		})
	}
}