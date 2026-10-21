package comments

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(
	mux *http.ServeMux,
	db *sql.DB,
	imageService ImageService,
) {
	repository := NewRepository(db)
	service := NewService(repository, imageService)
	handler := NewHandler(service)

	mux.HandleFunc("POST /comments/list", handler.Listar)
	mux.HandleFunc("POST /comments/create", handler.Criar)
	mux.HandleFunc("POST /comments/update", handler.Atualizar)
	mux.HandleFunc("POST /comments/delete", handler.Excluir)
}
