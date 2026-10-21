package schedule

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
// POST /horarios-disponiveis
// ============================================================

func (h *Handler) HorariosDisponiveis(w http.ResponseWriter, r *http.Request) {
	var req HorariosDisponiveisRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
		return
	}

	usaHorarios, horarios, err := h.service.BuscarHorarios(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			falha(w, http.StatusBadRequest, "salao_id, profissional_id e data são obrigatórios.")
		case errors.Is(err, ErrDataInvalida):
			falha(w, http.StatusBadRequest, "Data inválida.")
		case errors.Is(err, ErrProfissionalNaoEncontrado):
			falha(w, http.StatusNotFound, "Profissional não encontrado.")
		default:
			log.Println("erro ao buscar horários:", err)
			falha(w, http.StatusInternalServerError, "Erro ao buscar horários disponíveis.")
		}
		return
	}

	escreverJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"usa_horarios": usaHorarios,
			"horarios":     horarios,
		},
	})
}

// ============================================================
// POST /schedule-page
// ============================================================

func (h *Handler) SchedulePage(w http.ResponseWriter, r *http.Request) {
	var req SchedulePageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
		return
	}

	data, err := h.service.SchedulePage(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			falha(w, http.StatusBadRequest, "salao_id é obrigatório.")
		case errors.Is(err, ErrDataInvalida):
			falha(w, http.StatusBadRequest, "Data inválida.")
		case errors.Is(err, ErrSalaoNaoEncontrado):
			falha(w, http.StatusNotFound, "Salão não encontrado ou inativo.")
		case errors.Is(err, ErrProfissionalNaoEncontrado):
			falha(w, http.StatusNotFound, "Profissional não encontrado.")
		default:
			log.Println("erro ao carregar tela de agendamento:", err)
			falha(w, http.StatusInternalServerError, "Erro ao carregar tela de agendamento.")
		}
		return
	}

	escreverJSON(w, http.StatusOK, SchedulePageResponse{
		Success: true,
		Data:    data,
	})
}

// ============================================================
// POST /agendar
// ============================================================

func (h *Handler) Agendar(w http.ResponseWriter, r *http.Request) {
	var req AgendarRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
		return
	}

	agendamentoID, err := h.service.Agendar(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			falha(w, http.StatusBadRequest, "Informe salão, usuário, profissional, serviço, data e horário.")
		case errors.Is(err, ErrDataInvalida):
			falha(w, http.StatusBadRequest, "Data inválida ou não disponível.")
		case errors.Is(err, ErrHoraInvalida):
			falha(w, http.StatusBadRequest, "Horário inválido.")
		case errors.Is(err, ErrSalaoNaoEncontrado):
			falha(w, http.StatusNotFound, "Salão não encontrado ou inativo.")
		case errors.Is(err, ErrClienteNaoEncontrado):
			falha(w, http.StatusNotFound, "Cliente não encontrado.")
		case errors.Is(err, ErrProfissionalNaoEncontrado):
			falha(w, http.StatusNotFound, "Profissional não encontrado.")
		case errors.Is(err, ErrServicoNaoEncontrado):
			falha(w, http.StatusBadRequest, "Um ou mais serviços não estão disponíveis para este profissional.")
		case errors.Is(err, ErrDiaBloqueado):
			falha(w, http.StatusConflict, "Este dia está bloqueado para o profissional.")
		case errors.Is(err, ErrHorarioBloqueado):
			falha(w, http.StatusConflict, "Este horário está bloqueado para o profissional.")
		case errors.Is(err, ErrHorarioIndisponivel):
			falha(w, http.StatusConflict, "Horário indisponível. Escolha outro horário.")
		case errors.Is(err, ErrProfissionalSemHorarios):
			falha(w, http.StatusBadRequest, "Este profissional não trabalha com horários fixos.")
		default:
			log.Println("erro ao criar agendamento:", err)
			falha(w, http.StatusInternalServerError, "Erro ao criar agendamento.")
		}
		return
	}

	escreverJSON(w, http.StatusOK, AgendarResponse{
		Success: true,
		Message: "Agendamento realizado com sucesso.",
		Data: &struct {
			AgendamentoID int64 `json:"agendamento_id"`
		}{
			AgendamentoID: agendamentoID,
		},
	})
}

