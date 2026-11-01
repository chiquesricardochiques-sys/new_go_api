package agendamentoavulso

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
			falha(w, http.StatusBadRequest, "salão, profissional e data são obrigatórios.")
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
			falha(w, http.StatusBadRequest, "salão é obrigatório.")
		case errors.Is(err, ErrDataInvalida):
			falha(w, http.StatusBadRequest, "Data inválida.")
		case errors.Is(err, ErrSalaoNaoEncontrado):
			falha(w, http.StatusNotFound, "Salão não encontrado ou inativo.")
		case errors.Is(err, ErrProfissionalNaoEncontrado):
			falha(w, http.StatusNotFound, "Profissional não encontrado.")
		default:
			log.Println("erro ao carregar tela de agendamento avulso:", err)
			falha(w, http.StatusInternalServerError, "Erro ao carregar agendamento.")
		}
		return
	}

	escreverJSON(w, http.StatusOK, SchedulePageResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handler) ServicosProfissional(w http.ResponseWriter, r *http.Request) {
	var req ServicosProfissionalRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
		return
	}

	servicos, err := h.service.ServicosDoProfissional(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			falha(w, http.StatusBadRequest, "salão e profissional são obrigatórios.")
		case errors.Is(err, ErrProfissionalNaoEncontrado):
			falha(w, http.StatusNotFound, "Profissional não encontrado.")
		default:
			log.Println("erro ao buscar serviços:", err)
			falha(w, http.StatusInternalServerError, "Erro ao buscar serviços.")
		}
		return
	}

	escreverJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"servicos": servicos,
		},
	})
}

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
			falha(w, http.StatusBadRequest, "Informe nome, telefone, salão, profissional, serviço, data e horário.")
		case errors.Is(err, ErrDataInvalida):
			falha(w, http.StatusBadRequest, "Data inválida ou não disponível.")
		case errors.Is(err, ErrHoraInvalida):
			falha(w, http.StatusBadRequest, "Horário inválido.")
		case errors.Is(err, ErrSalaoNaoEncontrado):
			falha(w, http.StatusNotFound, "Salão não encontrado ou inativo.")
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
			log.Println("erro ao criar agendamento avulso:", err)
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


func (h *Handler) Profissionais(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req struct {
		SalaoID int64 `json:"salao_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		falha(w, http.StatusBadRequest, "JSON inválido.")
		return
	}

	profissionais, err := h.service.ListarProfissionais(
		r.Context(),
		req.SalaoID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrDadosInvalidos):
			falha(w, http.StatusBadRequest, "Salão obrigatório.")
		case errors.Is(err, ErrSalaoNaoEncontrado):
			falha(w, http.StatusNotFound, "Salão não encontrado ou inativo.")
		default:
			log.Println("erro ao listar profissionais:", err)
			falha(w, http.StatusInternalServerError, "Erro ao listar profissionais.")
		}
		return
	}

	escreverJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": profissionais,
	})
}
