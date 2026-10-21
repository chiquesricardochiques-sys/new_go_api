package auth

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	mux.HandleFunc("/api/admin/login", handler.LoginAdmin)
	mux.HandleFunc("/api/profissional/login", handler.LoginProfissional)
}
