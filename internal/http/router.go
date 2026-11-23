
package httpapi

import (
	"database/sql"
	"net/http"

	"github.com/ricardo-platform/salao-usuarios-api/internal/http/handlers"
	"github.com/ricardo-platform/salao-usuarios-api/internal/http/middleware"

	adminAdmins "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/admins"
	adminAvisos "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/avisos"
	adminConfProf "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/conf_prof"
	adminDashboard "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/dashboard"
	adminMysaloes "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/mysaloes"
	adminProfessionals "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/professionals"
	adminSalao "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/salao"
	adminSalary "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/salary"
	adminSchedule "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/agendamentoAvulso"
	adminServices "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/services"
	adminTextosSalao "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/textos_salao"
	professionalSchedule "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/professional_schedule"
	verAgendamentosAvulso "github.com/ricardo-platform/salao-usuarios-api/internal/modules/admin/verAgendamentosAvulso"

	auth "github.com/ricardo-platform/salao-usuarios-api/internal/modules/auth"
	clienteAuth "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/auth"
	clienteClientHome "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/client_home"
	clienteDadosAgendamentos "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/dadosAgendamentos"
	clienteHome "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/home"
	clienteSaloes "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/saloes"
	mySchedule "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/my-schedule"
	schedule "github.com/ricardo-platform/salao-usuarios-api/internal/modules/cliente/schedule"

	controleStatus "github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/controleStatus"
	profissionalGrade "github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/gradeHorarios"
	profissionalOps "github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/operations"
	profissional "github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/profissional"
	profissionalCodeEnt "github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/codeEnt"
	addImg "github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/addImg"

	dashboardPainel "github.com/ricardo-platform/salao-usuarios-api/internal/modules/dashboard/painel"
)

// NewRouter registra as rotas da API.
func NewRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// =====================================================
	// ADMINISTRADORES
	// =====================================================

	adminAdmins.RegisterRoutes(mux, db)
	adminMysaloes.RegisterRoutes(mux, db)
	adminAvisos.RegisterRoutes(mux, db)

	// =====================================================
	// PROFISSIONAIS
	// =====================================================

	profissional.RegisterRoutes(mux, db)
	addImg.RegisterRoutes(mux, db)

	// Entrada e localização de salão pelo código do painel.
	profissionalCodeEnt.RegisterRoutes(mux, db)

	// =====================================================
	// CLIENTES
	// =====================================================

	clienteAuth.RegisterRoutes(mux, db)
	clienteHome.RegisterRoutes(mux, db)
	clienteClientHome.RegisterRoutes(mux, db)
	schedule.RegisterRoutes(mux, db)
	mySchedule.RegisterRoutes(mux, db)
	clienteDadosAgendamentos.RegisterRoutes(mux, db)
	clienteSaloes.RegisterRoutes(mux, db)

	// =====================================================
	// SAÚDE DA API
	// =====================================================

	healthHandler := handlers.NewHealthHandler(db)
	mux.Handle("GET /health", healthHandler)

	// =====================================================
	// AUTENTICAÇÃO
	// =====================================================

	auth.RegisterRoutes(mux, db)

	// =====================================================
	// PAINEL ADMINISTRATIVO
	// =====================================================

	adminDashboard.RegisterRoutes(mux, db)

	// =====================================================
	// GERENCIAMENTO DE PROFISSIONAIS
	// =====================================================

	adminProfessionals.RegisterRoutes(mux, db)
	profissionalGrade.RegisterRoutes(mux, db)

	// =====================================================
	// CONFIGURAÇÕES DOS PROFISSIONAIS
	// =====================================================

	adminConfProf.RegisterRoutes(mux, db)

	// =====================================================
	// GERENCIAMENTO DO SALÃO
	// =====================================================

	adminSalao.RegisterRoutes(mux, db)

	// =====================================================
	// TEXTOS E INFORMAÇÕES DO SALÃO
	// =====================================================

	adminTextosSalao.RegisterRoutes(mux, db)

	// =====================================================
	// AGENDAMENTOS MANUAIS
	// =====================================================

	adminSchedule.RegisterRoutes(mux, db)
	verAgendamentosAvulso.RegisterRoutes(mux, db)

	// =====================================================
	// PAINEL DO CABELEIREIRO
	// =====================================================

	dashboardPainel.RegisterRoutes(mux, db)

	// =====================================================
	// GERENCIAMENTO DE SERVIÇOS
	// =====================================================

	servicesRepository := adminServices.NewRepository(db)
	servicesService := adminServices.NewService(servicesRepository)
	servicesHandler := adminServices.NewHandler(servicesService)

	adminServices.RegisterRoutes(mux, servicesHandler)

	// =====================================================
	// GRADE SEMANAL DOS PROFISSIONAIS
	// =====================================================

	professionalSchedule.RegisterRoutes(mux, db)

	// =====================================================
	// STATUS DOS CORTES
	// =====================================================

	controleStatus.RegisterRoutes(mux, db)

	// =====================================================
	// SALÁRIOS E RENDIMENTOS
	// =====================================================

	adminSalary.RegisterRoutes(mux, db)

	// =====================================================
	// OPERAÇÕES DO PROFISSIONAL
	// =====================================================

	profissionalOps.RegisterRoutes(mux, db)

	// =====================================================
	// CORS
	// =====================================================

	return middleware.CORS(mux)
}
