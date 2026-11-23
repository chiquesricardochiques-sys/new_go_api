package admins

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	// A rota addRelations foi removida: o vínculo agora nasce ao criar o salão
	// (módulo mysaloes), nunca por chamada direta com ids vindos do cliente.
	mux.HandleFunc(
		"POST /api/admin/admins/verify",
		handler.Verify,
	)
}