package blocks

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	mux.HandleFunc("POST /blocks/list", handler.Listar)
	mux.HandleFunc("POST /blocks/day/add", handler.AdicionarDia)
	mux.HandleFunc("POST /blocks/hour/add", handler.AdicionarHora)
	mux.HandleFunc("POST /blocks/day/remove", handler.RemoverDia)
	mux.HandleFunc("POST /blocks/hour/remove", handler.RemoverHora)
}
