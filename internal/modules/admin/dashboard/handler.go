
package dashboard

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		responder(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"message": "Método não permitido.",
		})
		return
	}

	idSalao := int64(0)
	valorQuery := strings.TrimSpace(r.URL.Query().Get("id_salao"))

	if valorQuery != "" {
		id, err := strconv.ParseInt(valorQuery, 10, 64)
		if err != nil || id <= 0 {
			responder(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"message": "id_salao inválido.",
			})
			return
		}

		idSalao = id
	} else {
		var body struct {
			IDSalao int64 `json:"id_salao"`
		}

		if r.Body != nil {
			err := json.NewDecoder(r.Body).Decode(&body)

			if err != nil && !errors.Is(err, io.EOF) {
				responder(w, http.StatusBadRequest, map[string]any{
					"success": false,
					"message": "Corpo JSON inválido.",
				})
				return
			}
		}

		idSalao = body.IDSalao
	}

	if idSalao <= 0 {
		responder(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "id_salao é obrigatório.",
		})
		return
	}

	resumo, err := h.service.Resumo(r.Context(), idSalao)
	if err != nil {
		if errors.Is(err, ErrSalaoInvalido) {
			responder(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		log.Printf("Erro ao carregar dashboard: %v", err)

		responder(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Erro ao carregar dashboard.",
		})
		return
	}

	responder(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    resumo,
	})
}

func responder(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
