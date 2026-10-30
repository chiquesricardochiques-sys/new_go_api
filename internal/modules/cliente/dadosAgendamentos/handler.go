package dadosagendamentos

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
// POST /dados-agendamento
// body: { "salao_id": 1, "agendamento_id": 8, "id_cliente": 1 }
// ============================================================

func (h *Handler) Buscar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req DadosRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
		return
	}

	dados, err := h.service.Buscar(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			falha(w, http.StatusBadRequest, "salao_id, agendamento_id e id_cliente são obrigatórios.")
		case errors.Is(err, ErrNaoEncontrado):
			falha(w, http.StatusNotFound, "Agendamento não encontrado.")
		case errors.Is(err, ErrNaoPendente):
			falha(w, http.StatusConflict, "Só é possível ver os detalhes de agendamentos pendentes.")
		default:
			log.Println("erro ao buscar dados do agendamento:", err)
			falha(w, http.StatusInternalServerError, "Erro ao carregar os dados do agendamento.")
		}
		return
	}

	escreverJSON(w, http.StatusOK, DadosResponse{
		Success: true,
		Data:    dados,
	})
}