package professional_schedule

import (
	"encoding/json"
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

type response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	payload response,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

func decodeJSON(
	r *http.Request,
	target interface{},
) error {
	return json.NewDecoder(r.Body).Decode(target)
}

func (h *Handler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req GetScheduleRequest

	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{
			Success: false,
			Message: "JSON inválido",
		})
		return
	}

	if req.SalaoID == 0 || req.ProfissionalID == 0 {
		writeJSON(w, http.StatusBadRequest, response{
			Success: false,
			Message: "salao_id e profissional_id são obrigatórios",
		})
		return
	}

	grade, err := h.service.ObterGrade(
		req.SalaoID,
		req.ProfissionalID,
	)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, response{
		Success: true,
		Data: map[string]interface{}{
			"grade": grade,
		},
	})
}

func (h *Handler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req UpdateScheduleRequest

	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{
			Success: false,
			Message: "JSON inválido",
		})
		return
	}

	if req.SalaoID == 0 || req.ProfissionalID == 0 {
		writeJSON(w, http.StatusBadRequest, response{
			Success: false,
			Message: "salao_id e profissional_id são obrigatórios",
		})
		return
	}

	if req.Grade == nil {
		req.Grade = []ScheduleDay{}
	}

	if err := h.service.SalvarGrade(
		req.SalaoID,
		req.ProfissionalID,
		req.Grade,
	); err != nil {
		writeJSON(w, http.StatusBadRequest, response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, response{
		Success: true,
		Message: "Grade semanal salva com sucesso",
	})
}

func (h *Handler) Reset(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req ResetScheduleRequest

	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{
			Success: false,
			Message: "JSON inválido",
		})
		return
	}

	if req.SalaoID == 0 || req.ProfissionalID == 0 {
		writeJSON(w, http.StatusBadRequest, response{
			Success: false,
			Message: "salao_id e profissional_id são obrigatórios",
		})
		return
	}

	if err := h.service.ResetarGrade(
		req.SalaoID,
		req.ProfissionalID,
	); err != nil {
		writeJSON(w, http.StatusBadRequest, response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, response{
		Success: true,
		Message: "Grade semanal resetada com sucesso",
	})
}