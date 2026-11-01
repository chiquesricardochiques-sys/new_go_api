package verAgendamentosAvulso

import (
    "database/sql"
    "net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
    repository := NewRepository(db)
    service := NewService(repository)
    handler := NewHandler(service)

    mux.HandleFunc(
        "POST /agendamentos-avulsos/list",
        handler.Listar,
    )

    mux.HandleFunc(
        "POST /agendamentos-avulsos/cancel",
        handler.Cancelar,
    )
}