package services

import "net/http"

func RegisterRoutes(
	mux *http.ServeMux,
	handler *Handler,
) {
	// ========================================================
	// SERVIÇOS GLOBAIS
	// ========================================================

	mux.HandleFunc(
		"/services/global/list",
		handler.ListGlobalServices,
	)

	// ========================================================
	// SERVIÇOS DO SALÃO
	// ========================================================

	mux.HandleFunc(
		"/services/salon/list",
		handler.ListSalonServices,
	)

	mux.HandleFunc(
		"/services/salon/add",
		handler.AddSalonService,
	)

	mux.HandleFunc(
		"/services/salon/deactivate",
		handler.DeactivateSalonService,
	)

	mux.HandleFunc(
		"/services/salon/unlink",
		handler.UnlinkSalonService,
	)

	// Busca individual.
	mux.HandleFunc(
		"/services/get",
		handler.GetService,
	)

	// ========================================================
	// PROFISSIONAIS
	// ========================================================

	mux.HandleFunc(
		"/services/professionals/list",
		handler.ListProfessionals,
	)

	mux.HandleFunc(
		"/services/professionals/by-service",
		handler.ListProfessionalsByService,
	)

	// ========================================================
	// SERVIÇOS DO PROFISSIONAL
	// ========================================================

	// Checklist completo.
	mux.HandleFunc(
		"/services/professionals/services/list",
		handler.ListProfessionalServices,
	)

	// Somente os vinculados.
	mux.HandleFunc(
		"/services/professionals/services/linked/list",
		handler.ListLinkedProfessionalServices,
	)

	mux.HandleFunc(
		"/services/professionals/services/add",
		handler.AddProfessionalService,
	)

	mux.HandleFunc(
		"/services/professionals/services/remove",
		handler.RemoveProfessionalService,
	)
}