package horarios

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	mux.HandleFunc("POST /horarios/list", handler.Listar)
	mux.HandleFunc("POST /horarios/create", handler.Criar)
	mux.HandleFunc("POST /horarios/delete", handler.Excluir)
}
