package mysaloes

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	mux.HandleFunc(
		"POST /api/admin/meus-saloes/listar",
		handler.Listar,
	)

	mux.HandleFunc(
		"POST /api/admin/meus-saloes/criar",
		handler.Criar,
	)

	mux.HandleFunc(
		"POST /api/admin/meus-saloes/atualizar",
		handler.Atualizar,
	)

	mux.HandleFunc(
		"POST /api/admin/meus-saloes/entrar",
		handler.Entrar,
	)

	mux.HandleFunc(
		"POST /api/admin/meus-saloes/imagens/adicionar",
		handler.AdicionarImagem,
	)

	mux.HandleFunc(
		"POST /api/admin/meus-saloes/imagens/atualizar",
		handler.AtualizarImagem,
	)

	mux.HandleFunc(
		"POST /api/admin/meus-saloes/imagens/listar",
		handler.ListarImagens,
	)

	mux.HandleFunc(
		"POST /api/admin/meus-saloes/atualizacao",
		handler.GetAtualizacao,
	)
}