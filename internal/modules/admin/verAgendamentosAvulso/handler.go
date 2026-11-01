package verAgendamentosAvulso

import (
    "encoding/json"
    "errors"
    "net/http"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func responderJSON(
    w http.ResponseWriter,
    status int,
    dados any,
) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(dados)
}

func (h *Handler) Listar(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        responderJSON(w, http.StatusMethodNotAllowed, map[string]any{
            "sucesso": false,
            "mensagem": "Método não permitido",
        })
        return
    }

    var req FiltroAgendamentosAvulsos

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        responderJSON(w, http.StatusBadRequest, map[string]any{
            "sucesso": false,
            "mensagem": "Dados inválidos",
        })
        return
    }

    agendamentos, err := h.service.Listar(r.Context(), req)
    if err != nil {
        responderJSON(w, http.StatusBadRequest, map[string]any{
            "sucesso": false,
            "mensagem": err.Error(),
        })
        return
    }

    responderJSON(w, http.StatusOK, map[string]any{
        "sucesso": true,
        "data": req.Data,
        "total": len(agendamentos),
        "agendamentos": agendamentos,
    })
}

func (h *Handler) Cancelar(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        responderJSON(w, http.StatusMethodNotAllowed, map[string]any{
            "sucesso": false,
            "mensagem": "Método não permitido",
        })
        return
    }

    var req CancelarAgendamentoRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        responderJSON(w, http.StatusBadRequest, map[string]any{
            "sucesso": false,
            "mensagem": "Dados inválidos",
        })
        return
    }

    err := h.service.Cancelar(r.Context(), req)

    if err != nil {
        switch {
        case errors.Is(err, ErrAgendamentoNaoEncontrado):
            responderJSON(w, http.StatusNotFound, map[string]any{
                "sucesso": false,
                "mensagem": err.Error(),
            })

        case errors.Is(err, ErrAgendamentoJaCancelado):
            responderJSON(w, http.StatusConflict, map[string]any{
                "sucesso": false,
                "mensagem": err.Error(),
            })

        default:
            responderJSON(w, http.StatusBadRequest, map[string]any{
                "sucesso": false,
                "mensagem": err.Error(),
            })
        }

        return
    }

    responderJSON(w, http.StatusOK, CancelarAgendamentoResponse{
        Sucesso: true,
        Mensagem: "Agendamento avulso cancelado com sucesso",
    })
}