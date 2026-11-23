package saloes

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(
	mux *http.ServeMux,
	db *sql.DB,
) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	mux.HandleFunc("POST /saloes/buscar", handler.Buscar)
	mux.HandleFunc("POST /saloes/por-slug", handler.PorSlug)
	mux.HandleFunc("POST /saloes/detalhes", handler.Detalhes)
	mux.HandleFunc("POST /saloes/avaliar", handler.Avaliar)
	mux.HandleFunc("POST /saloes/cidades", handler.Cidades)
}