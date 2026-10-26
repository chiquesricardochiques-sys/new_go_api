package addimg

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ============================================================================
// POST /profissional/foto
// body: {"id_salao":1,"profissional_id":3,"foto_url":"https://...","foto_public_id":"..."}
// ============================================================================

func (h *Handler) Adicionar(w http.ResponseWriter, r *http.Request) {
	var req FotoRequest
	if err := lerJSON(w, r, &req); err != nil {
		h.responderErro(w, err)
		return
	}

	foto, err := h.service.Adicionar(r.Context(), req)
	if err != nil {
		h.responderErro(w, err)
		return
	}

	escreverJSON(w, http.StatusCreated, FotoResponse{
		Success: true,
		Message: "Foto adicionada com sucesso.",
		Foto:    foto,
	})
}

// ============================================================================
// PUT /profissional/foto
// body: igual ao POST
// ============================================================================

func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	var req FotoRequest
	if err := lerJSON(w, r, &req); err != nil {
		h.responderErro(w, err)
		return
	}

	foto, anterior, err := h.service.Atualizar(r.Context(), req)
	if err != nil {
		h.responderErro(w, err)
		return
	}

	escreverJSON(w, http.StatusOK, FotoResponse{
		Success:  true,
		Message:  "Foto atualizada com sucesso.",
		Foto:     foto,
		Anterior: anterior,
	})
}

// ============================================================================
// DELETE /profissional/foto
// body: {"id_salao":1,"profissional_id":3}
// ============================================================================

func (h *Handler) Remover(w http.ResponseWriter, r *http.Request) {
	var req RemoverRequest
	if err := lerJSON(w, r, &req); err != nil {
		h.responderErro(w, err)
		return
	}

	anterior, err := h.service.Remover(r.Context(), req)
	if err != nil {
		h.responderErro(w, err)
		return
	}

	escreverJSON(w, http.StatusOK, FotoResponse{
		Success:  true,
		Message:  "Foto removida com sucesso.",
		Anterior: anterior,
	})
}

// ============================================================================
// APOIO
// ============================================================================

func lerJSON(w http.ResponseWriter, r *http.Request, destino any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if err := json.NewDecoder(r.Body).Decode(destino); err != nil {
		return fmt.Errorf("%w: corpo JSON inválido", ErrDadosInvalidos)
	}
	return nil
}

func (h *Handler) responderErro(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrDadosInvalidos):
		escreverJSON(w, http.StatusBadRequest, FotoResponse{Success: false, Message: err.Error()})
	case errors.Is(err, ErrNaoEncontrado), errors.Is(err, ErrSemFoto):
		escreverJSON(w, http.StatusNotFound, FotoResponse{Success: false, Message: err.Error()})
	case errors.Is(err, ErrJaPossuiFoto):
		escreverJSON(w, http.StatusConflict, FotoResponse{Success: false, Message: err.Error()})
	default:
		log.Printf("[profissional/addImg] erro interno: %v", err)
		escreverJSON(w, http.StatusInternalServerError, FotoResponse{Success: false, Message: "Erro interno."})
	}
}

func escreverJSON(w http.ResponseWriter, status int, corpo FotoResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(corpo)
}