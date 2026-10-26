package conf_prof

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// =========================================================
// ADMIN
// =========================================================
//
// GET
// /admin/configuracao-profissional/rendimento/{profissionalId}
//
// Se não existir configuração, o Service cria automaticamente
// como permitido = true.
//

func (h *Handler) BuscarPodeVerRendimento(
	w http.ResponseWriter,
	r *http.Request,
) {
	profissionalID, ok := profissionalIDDaURL(r)

	if !ok {
		writeResponse(
			w,
			http.StatusBadRequest,
			Response{
				Success: false,
				Message: "ID do profissional inválido.",
			},
		)
		return
	}

	permitido, err := h.service.PodeVerRendimento(
		r.Context(),
		profissionalID,
	)

	if err != nil {
		h.tratarErro(w, err)
		return
	}

	writeResponse(
		w,
		http.StatusOK,
		Response{
			Success: true,
			Data: PodeVerRendimento{
				ProfissionalID: profissionalID,
				Permitido:      permitido,
			},
		},
	)
}

// =========================================================
// ADMIN
// =========================================================
//
// PUT
// /admin/configuracao-profissional/rendimento/{profissionalId}
//
// Body:
//
// {
//   "permitido": true
// }
//
// ou
//
// {
//   "permitido": false
// }
//

func (h *Handler) DefinirPodeVerRendimento(
	w http.ResponseWriter,
	r *http.Request,
) {
	profissionalID, ok := profissionalIDDaURL(r)

	if !ok {
		writeResponse(
			w,
			http.StatusBadRequest,
			Response{
				Success: false,
				Message: "ID do profissional inválido.",
			},
		)
		return
	}

	var req struct {
		Permitido *bool `json:"permitido"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeResponse(
			w,
			http.StatusBadRequest,
			Response{
				Success: false,
				Message: "JSON inválido.",
			},
		)
		return
	}

	if req.Permitido == nil {
		writeResponse(
			w,
			http.StatusBadRequest,
			Response{
				Success: false,
				Message: "Informe o campo permitido.",
			},
		)
		return
	}

	err := h.service.DefinirPodeVerRendimento(
		r.Context(),
		profissionalID,
		*req.Permitido,
	)

	if err != nil {
		h.tratarErro(w, err)
		return
	}

	writeResponse(
		w,
		http.StatusOK,
		Response{
			Success: true,
			Data: PodeVerRendimento{
				ProfissionalID: profissionalID,
				Permitido:      *req.Permitido,
			},
			Message: "Configuração atualizada com sucesso.",
		},
	)
}

// =========================================================
// PROFISSIONAL
// =========================================================
//
// Temporariamente usando profissional 3.
//
// Depois vamos trocar isso pelo profissional vindo da
// autenticação.
//

func (h *Handler) BuscarPodeVerRendimentoProfissional(
	w http.ResponseWriter,
	r *http.Request,
) {
	const profissionalID int64 = 3

	permitido, err := h.service.PodeVerRendimento(
		r.Context(),
		profissionalID,
	)

	if err != nil {
		h.tratarErro(w, err)
		return
	}

	writeResponse(
		w,
		http.StatusOK,
		Response{
			Success: true,
			Data: PodeVerRendimento{
				ProfissionalID: profissionalID,
				Permitido:      permitido,
			},
		},
	)
}

// =========================================================
// ERROS
// =========================================================

func (h *Handler) tratarErro(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(
		err,
		ErrProfissionalInvalido,
	):
		writeResponse(
			w,
			http.StatusBadRequest,
			Response{
				Success: false,
				Message: "Profissional inválido.",
			},
		)

	default:
		writeResponse(
			w,
			http.StatusInternalServerError,
			Response{
				Success: false,
				Message: "Erro ao consultar configuração do profissional.",
			},
		)
	}
}

// =========================================================
// URL
// =========================================================

func profissionalIDDaURL(
	r *http.Request,
) (int64, bool) {

	idTexto := strings.TrimSpace(
		r.PathValue("profissionalId"),
	)

	if idTexto == "" {
		return 0, false
	}

	id, err := strconv.ParseInt(
		idTexto,
		10,
		64,
	)

	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}

// =========================================================
// RESPONSE
// =========================================================

func writeResponse(
	w http.ResponseWriter,
	status int,
	body Response,
) {
	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(body)
}