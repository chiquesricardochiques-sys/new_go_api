package dashboard

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Summary(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodGet &&
		r.Method != http.MethodPost {

		w.Header().Set(
			"Allow",
			http.MethodGet+", "+http.MethodPost,
		)

		responder(
			w,
			http.StatusMethodNotAllowed,
			map[string]any{
				"success": false,
				"message": "Método não permitido.",
			},
		)

		return
	}

	// --------------------------------------------------------
	// id_salao
	// --------------------------------------------------------

	idSalao := int64(0)

	valorQuery := strings.TrimSpace(
		r.URL.Query().Get("id_salao"),
	)

	if valorQuery != "" {

		id, err := strconv.ParseInt(
			valorQuery,
			10,
			64,
		)

		if err != nil || id <= 0 {

			responder(
				w,
				http.StatusBadRequest,
				map[string]any{
					"success": false,
					"message": "id_salao inválido.",
				},
			)

			return
		}

		idSalao = id
	}

	// --------------------------------------------------------
	// Compatibilidade com POST antigo
	// --------------------------------------------------------

	if idSalao <= 0 &&
		r.Method == http.MethodPost &&
		r.Body != nil {

		var body struct {
			IDSalao int64 `json:"id_salao"`
		}

		err := json.NewDecoder(
			r.Body,
		).Decode(&body)

		if err == nil {
			idSalao = body.IDSalao
		}
	}

	// --------------------------------------------------------
	// Validação
	// --------------------------------------------------------

	if idSalao <= 0 {

		responder(
			w,
			http.StatusBadRequest,
			map[string]any{
				"success": false,
				"message": "id_salao é obrigatório.",
			},
		)

		return
	}

	// --------------------------------------------------------
	// Dashboard
	// --------------------------------------------------------

	resumo, err := h.service.Resumo(
		r.Context(),
		idSalao,
	)

	if err != nil {

		if errors.Is(
			err,
			ErrSalaoInvalido,
		) {

			responder(
				w,
				http.StatusBadRequest,
				map[string]any{
					"success": false,
					"message": err.Error(),
				},
			)

			return
		}

		log.Printf(
			"Erro ao carregar dashboard: %v",
			err,
		)

		responder(
			w,
			http.StatusInternalServerError,
			map[string]any{
				"success": false,
				"message": "Erro ao carregar dashboard.",
			},
		)

		return
	}

	responder(
		w,
		http.StatusOK,
		map[string]any{
			"success": true,
			"data":    resumo,
		},
	)
}

func responder(
	w http.ResponseWriter,
	status int,
	data any,
) {

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}