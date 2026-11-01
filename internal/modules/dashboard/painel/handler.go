package painel

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	service *Service
	limite  *limitador
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
		// 10 códigos errados a cada 10 minutos, por IP
		limite: novoLimitador(10, 10*time.Minute),
	}
}

func escreverJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func falha(w http.ResponseWriter, status int, mensagem string) {
	escreverJSON(w, status, map[string]any{
		"success": false,
		"message": mensagem,
	})
}

// ============================================================
// POST /painel/agenda
// body: { "codigo": "#0815332", "data": "2026-10-31", "profissional_id": 0 }
//
// Valida o código, e devolve o salão, os profissionais e a grade
// de agendamentos do dia. "data" e "profissional_id" são opcionais.
// ============================================================

func (h *Handler) Agenda(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req AgendaRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
		return
	}

	chave := ipDe(r)

	if h.limite.bloqueado(chave) {
		w.Header().Set("Retry-After", "600")

		falha(
			w,
			http.StatusTooManyRequests,
			"Muitas tentativas. Aguarde alguns minutos e tente de novo.",
		)
		return
	}

	data, err := h.service.Agenda(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			falha(w, http.StatusBadRequest, "Informe o código do salão.")
		case errors.Is(err, ErrDataInvalida):
			falha(w, http.StatusBadRequest, "Data inválida.")
		case errors.Is(err, ErrCodigoInvalido):
			h.limite.registrarFalha(chave)
			falha(w, http.StatusNotFound, "Código não encontrado.")
		default:
			log.Println("erro ao carregar painel:", err)
			falha(w, http.StatusInternalServerError, "Erro ao carregar o painel.")
		}
		return
	}

	escreverJSON(w, http.StatusOK, AgendaResponse{
		Success: true,
		Data:    data,
	})
}