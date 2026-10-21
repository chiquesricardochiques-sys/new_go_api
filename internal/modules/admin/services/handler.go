package services

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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(
	w http.ResponseWriter,
	status int,
	err error,
) {
	writeJSON(
		w,
		status,
		response{
			Success: false,
			Message: err.Error(),
		},
	)
}

func decodeJSON(
	r *http.Request,
	target interface{},
) error {

	if r.Body == nil {
		return nil
	}

	return json.NewDecoder(r.Body).Decode(target)
}

// ============================================================
// SERVIÇOS GLOBAIS
// ============================================================

func (h *Handler) ListGlobalServices(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	servicos, err := h.service.ListGlobalServices(
		r.Context(),
	)

	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Data: map[string]interface{}{
				"servicos": servicos,
			},
		},
	)
}

// ============================================================
// SERVIÇOS DO SALÃO
// ============================================================

func (h *Handler) ListSalonServices(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	var req ListSalonServicesRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	servicos, err := h.service.ListSalonServices(
		r.Context(),
		req.SalaoID,
	)

	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Data: map[string]interface{}{
				"servicos": servicos,
			},
		},
	)
}

func (h *Handler) GetService(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	var req GetServiceRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	servico, err := h.service.GetSalonService(
		r.Context(),
		req.SalaoID,
		req.ServicoID,
	)

	if err != nil {
		if errors.Is(err, errors.New("serviço não encontrado")) {
			writeError(
				w,
				http.StatusNotFound,
				err,
			)
			return
		}

		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	if servico == nil {
		writeError(
			w,
			http.StatusNotFound,
			errors.New("serviço não encontrado"),
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Data: map[string]interface{}{
				"servico": servico,
			},
		},
	)
}

func (h *Handler) AddSalonService(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	var req AddSalonServiceRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	if err := h.service.AddSalonService(
		r.Context(),
		req,
	); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Message: "Serviço adicionado ao salão com sucesso.",
		},
	)
}

func (h *Handler) DeactivateSalonService(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	var req DeactivateSalonServiceRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	if err := h.service.DeactivateSalonService(
		r.Context(),
		req,
	); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Message: "Serviço desativado no salão.",
		},
	)
}

func (h *Handler) UnlinkSalonService(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	var req UnlinkSalonServiceRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	if err := h.service.UnlinkSalonService(
		r.Context(),
		req,
	); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Message: "Serviço desvinculado do salão.",
		},
	)
}

// ============================================================
// PROFISSIONAIS
// ============================================================

func (h *Handler) ListProfessionals(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	var req ListProfessionalsRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	profissionais, err := h.service.ListProfessionals(
		r.Context(),
		req.SalaoID,
	)

	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Data: map[string]interface{}{
				"profissionais": profissionais,
			},
		},
	)
}

func (h *Handler) ListProfessionalsByService(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	var req ListProfessionalsByServiceRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	profissionais, err := h.service.ListProfessionalsByService(
		r.Context(),
		req.SalaoID,
		req.ServicoID,
	)

	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Data: map[string]interface{}{
				"profissionais": profissionais,
			},
		},
	)
}

// ============================================================
// SERVIÇOS DO PROFISSIONAL
// ============================================================

func (h *Handler) ListProfessionalServices(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	var req ListProfessionalServicesRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	servicos, err := h.service.ListProfessionalServices(
		r.Context(),
		req.SalaoID,
		req.ProfissionalID,
	)

	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Data: map[string]interface{}{
				"servicos": servicos,
			},
		},
	)
}

func (h *Handler) ListLinkedProfessionalServices(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	var req ListLinkedProfessionalServicesRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	servicos, err := h.service.ListLinkedProfessionalServices(
		r.Context(),
		req.SalaoID,
		req.ProfissionalID,
	)

	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Data: map[string]interface{}{
				"servicos": servicos,
			},
		},
	)
}

func (h *Handler) AddProfessionalService(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	var req AddProfessionalServiceRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	if err := h.service.AddProfessionalService(
		r.Context(),
		req,
	); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Message: "Serviço vinculado ao profissional.",
		},
	)
}

func (h *Handler) RemoveProfessionalService(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			errors.New("método não permitido"),
		)
		return
	}

	var req RemoveProfessionalServiceRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	if err := h.service.RemoveProfessionalService(
		r.Context(),
		req,
	); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		response{
			Success: true,
			Message: "Serviço removido do profissional.",
		},
	)
}