package addimg

import (
	"database/sql"
	"net/http"
)

// RegisterRoutes monta repository -> service -> handler e registra as rotas
// da foto do profissional (colunas foto_url e foto_public_id).
//
//	POST   /profissional/foto   adiciona a primeira foto
//	PUT    /profissional/foto   altera a foto
//	DELETE /profissional/foto   remove a foto
//
// Usa o ServeMux do Go 1.22+ ("MÉTODO /caminho").
func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	mux.HandleFunc("POST /profissional/foto", handler.Adicionar)
	mux.HandleFunc("PUT /profissional/foto", handler.Atualizar)
	mux.HandleFunc("DELETE /profissional/foto", handler.Remover)
}