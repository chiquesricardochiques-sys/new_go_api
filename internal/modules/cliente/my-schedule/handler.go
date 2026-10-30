package myschedule

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

// POST /my-schedules  (agendamentos de hoje em diante)
func (h *Handler) MeusAgendamentos(w http.ResponseWriter, r *http.Request) {
	h.listar(w, r, CenarioMeusAgendamentos, "Erro ao carregar meus agendamentos.")
}

// POST /historico-agendamentos
func (h *Handler) Historico(w http.ResponseWriter, r *http.Request) {
	h.listar(w, r, CenarioHistorico, "Erro ao carregar histórico de agendamentos.")
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request, cenario, msgErro string) {
	var req ListarRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
		return
	}

	// a rota decide o cenário (ignora o que vier no body)
	req.Cenario = cenario

	data, err := h.service.Listar(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			falha(w, http.StatusBadRequest, "Cliente não identificado.")
		case errors.Is(err, ErrCenarioInvalido):
			falha(w, http.StatusBadRequest, "Cenário inválido.")
		case errors.Is(err, ErrStatusInvalido):
			falha(w, http.StatusBadRequest, "Status inválido.")
		default:
			log.Println("erro ao listar agendamentos:", err)
			falha(w, http.StatusInternalServerError, msgErro)
		}
		return
	}

	escreverJSON(w, http.StatusOK, ListarResponse{Success: true, Data: data})
}

// POST /deletar-agendamento  (na prática cancela: status = 'cancelado')
func (h *Handler) Cancelar(w http.ResponseWriter, r *http.Request) {
	var req CancelarRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
		return
	}

	if err := h.service.Cancelar(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			falha(w, http.StatusBadRequest, "agendamento_id e auth_user_id são obrigatórios.")
		case errors.Is(err, ErrAgendamentoNaoEncontrado):
			falha(w, http.StatusNotFound, "Agendamento não encontrado ou não permitido.")
		case errors.Is(err, ErrJaCancelado):
			falha(w, http.StatusConflict, "Este agendamento já foi cancelado.")
		case errors.Is(err, ErrNaoPodeCancelar):
			falha(w, http.StatusConflict, "Este agendamento não pode mais ser cancelado.")
		default:
			log.Println("erro ao cancelar agendamento:", err)
			falha(w, http.StatusInternalServerError, "Erro ao cancelar agendamento.")
		}
		return
	}

	escreverJSON(w, http.StatusOK, CancelarResponse{
		Success: true,
		Message: "Agendamento cancelado com sucesso!",
	})
}