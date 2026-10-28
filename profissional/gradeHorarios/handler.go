package gradehorarios

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func responder(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Sem DisallowUnknownFields: o app pode mandar campos a mais sem quebrar.
func lerJSON(w http.ResponseWriter, r *http.Request, destino any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if err := json.NewDecoder(r.Body).Decode(destino); err != nil {
		responder(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Corpo da requisição inválido.",
		})
		return false
	}

	return true
}

// comData: lê o JSON, chama o service e responde {"success":true,"message":..., "data":...}.
func comData[Req any, Res any](mensagem string, fn func(ctx context.Context, req Req) (Res, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Req

		if !lerJSON(w, r, &req) {
			return
		}

		resultado, err := fn(r.Context(), req)
		if err != nil {
			erro(w, err)
			return
		}

		corpo := map[string]any{
			"success": true,
			"data":    resultado,
		}
		if mensagem != "" {
			corpo["message"] = mensagem
		}

		responder(w, http.StatusOK, corpo)
	}
}

func erro(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrDadosInvalidos),
		errors.Is(err, ErrProfissionalSemHorarios),
		errors.Is(err, ErrHorarioPassado),
		errors.Is(err, ErrForaDaGrade):
		responder(w, http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})

	case errors.Is(err, ErrNaoEncontrado):
		responder(w, http.StatusNotFound, map[string]any{"success": false, "message": err.Error()})

	case errors.Is(err, ErrTemAgendamento),
		errors.Is(err, ErrJaBloqueado),
		errors.Is(err, ErrDiaBloqueado):
		responder(w, http.StatusConflict, map[string]any{"success": false, "message": err.Error()})

	default:
		log.Printf("[gradeHorarios] erro interno: %v", err)
		responder(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Erro ao processar a grade de horários.",
		})
	}
}