package avisos

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Listar(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)

		json.NewEncoder(w).Encode(ApiResponse{
			Success: false,
			Message: "Método não permitido",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request ListarAvisosRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(ApiResponse{
			Success: false,
			Message: "Dados da requisição inválidos",
		})
		return
	}

	if request.SalaoID <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(ApiResponse{
			Success: false,
			Message: "ID do salão inválido",
		})
		return
	}

	resultado, err := h.service.Listar(
		r.Context(),
		request.SalaoID,
		request.Historico,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(ApiResponse{
			Success: false,
			Message: "Não foi possível carregar os avisos",
		})
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(ApiResponse{
		Success: true,
		Data:    resultado,
	})
}