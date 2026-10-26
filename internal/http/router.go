package httpapi

import (
	"database/sql"
	"net/http"

	"github.com/ricardo-platform/salao-usuarios-api/internal/http/handlers"
	"github.com/ricardo-platform/salao-usuarios-api/internal/http/middleware"

	adminAdmins "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/admins"
	adminConfProf "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/conf_prof"
	adminDashboard "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/dashboard"
	adminProfessionals "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/professionals"
	addImg "github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/addImg"
	adminSalary "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/salary"
	adminServices "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/services"
	adminSalao "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/salao"
	adminTextosSalao "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/textos_salao"
	professionalSchedule "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/professional_schedule"

	clienteAuth "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/auth"

	auth "github.com/ricardo-platform/salao-usuarios-api/internal/modules/auth"
	profissional "github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/profissional"
	clienteHome "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/home"
	clienteClientHome "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/client_home"
	schedule "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/schedule"
	mySchedule "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/my-schedule"
	clienteDadosAgendamentos "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/dadosAgendamentos"
	profissionalOps "github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/operations"
	controleStatus "github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/controleStatus"
)

// NewRouter registra as rotas da API.
func NewRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// =====================================================
	// ADMINISTRADORES
	// =====================================================

	// Vínculos e validação de administradores.
	adminAdmins.RegisterRoutes(
		mux,
		db,
	)

	// =====================================================
	// PROFISSIONAIS
	// =====================================================

	// Vínculos e validação de profissionais.
	profissional.RegisterRoutes(
		mux,
		db,
	)
	// Adicionar, alterar e remover foto_url / foto_public_id.
	addImg.RegisterRoutes(
		mux,
		db,
	)

	// =====================================================
	// CLIENTES
	// =====================================================

	// Vínculo entre o usuário do AUTH e o cliente interno do banco.
	clienteAuth.RegisterRoutes(
		mux,
		db,
	)

	clienteHome.RegisterRoutes(mux, db)
	clienteClientHome.RegisterRoutes(mux, db)
	schedule.RegisterRoutes(mux, db)
	mySchedule.RegisterRoutes(mux, db)
	clienteDadosAgendamentos.RegisterRoutes(mux, db)

	// =====================================================
	// SAÚDE DA API
	// =====================================================

	// Saúde da API e conexão com o banco.
	healthHandler := handlers.NewHealthHandler(db)

	mux.Handle(
		"GET /health",
		healthHandler,
	)

	// =====================================================
	// AUTENTICAÇÃO
	// =====================================================

	// Autenticação de administradores e profissionais.
	auth.RegisterRoutes(
		mux,
		db,
	)

	// =====================================================
	// PAINEL ADMINISTRATIVO
	// =====================================================

	adminDashboard.RegisterRoutes(
		mux,
		db,
	)

	// =====================================================
	// GERENCIAMENTO DE PROFISSIONAIS
	// =====================================================

	adminProfessionals.RegisterRoutes(
		mux,
		db,
	)

	// =====================================================
	// CONFIGURAÇÕES DOS PROFISSIONAIS
	// =====================================================

	adminConfProf.RegisterRoutes(
		mux,
		db,
	)

	// =====================================================
	// GERENCIAMENTO DO SALÃO
	// =====================================================

	adminSalao.RegisterRoutes(
		mux,
		db,
	)

	// =====================================================
	// TEXTOS E INFORMAÇÕES DO SALÃO
	// =====================================================

	adminTextosSalao.RegisterRoutes(
		mux,
		db,
	)

	// =====================================================
	// GERENCIAMENTO DE SERVIÇOS
	// =====================================================

	servicesRepository := adminServices.NewRepository(
		db,
	)

	servicesService := adminServices.NewService(
		servicesRepository,
	)

	servicesHandler := adminServices.NewHandler(
		servicesService,
	)

	adminServices.RegisterRoutes(
		mux,
		servicesHandler,
	)

	// =====================================================
	// GRADE SEMANAL DOS PROFISSIONAIS
	// =====================================================

	professionalSchedule.RegisterRoutes(
		mux,
		db,
	)
	// rotas de estatus de cortes 
	controleStatus.RegisterRoutes(mux, db)
	// =====================================================
	// SALÁRIOS E RENDIMENTOS
	// =====================================================

	adminSalary.RegisterRoutes(
		mux,
		db,
	)

	// Operações do app do profissional (agenda, status, resumo, histórico, horários, perfil, faturamento).
	profissionalOps.RegisterRoutes(
		mux,
		db,
	)

	// =====================================================
	// CORS
	// =====================================================

	// CORS deve envolver o roteador inteiro.
	return middleware.CORS(
		mux,
	)


}