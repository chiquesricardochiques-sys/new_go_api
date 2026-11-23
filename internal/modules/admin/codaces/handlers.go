package codaces

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	controller *Controller
}

func NewHandler(controller *Controller) *Handler {
	return &Handler{
		controller: controller,
	}
}

// ------------------------------------------------------------
// BUSCAR CÓDIGO
// ------------------------------------------------------------

func (h *Handler) BuscarCodigo(
	w http.ResponseWriter,
	r *http.Request,
) {

	salaoID, err := obterSalaoID(r)
	if err != nil {
		responderErro(
			w,
			http.StatusBadRequest,
			"ID do salão inválido",
		)
		return
	}

	response, err := h.controller.BuscarCodigo(salaoID)

	if err != nil {
		switch {
		case errors.Is(err, ErrSalaoNaoEncontrado):
			responderErro(
				w,
				http.StatusNotFound,
				"Salão não encontrado",
			)

		case errors.Is(err, ErrCodigoNaoCadastrado):
			responderErro(
				w,
				http.StatusNotFound,
				"Salão ainda não possui código de vínculo",
			 )

		default:
			responderErro(
				w,
				http.StatusInternalServerError,
				"Não foi possível consultar o código",
			)
		}

		return
	}

	responderJSON(w, http.StatusOK, response)
}

// ------------------------------------------------------------
// INSERIR CÓDIGO
// ------------------------------------------------------------

func (h *Handler) InserirCodigo(
	w http.ResponseWriter,
	r *http.Request,
) {

	salaoID, err := obterSalaoID(r)
	if err != nil {
		responderErro(
			w,
			http.StatusBadRequest,
			"ID do salão inválido",
		)
		return
	}

	var request CodigoRequest

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		responderErro(
			w,
			http.StatusBadRequest,
			"Dados inválidos",
		)
		return
	}

	codigo := strings.TrimSpace(request.Codigo)

	err = h.controller.InserirCodigo(salaoID, codigo)

	if err != nil {
		h.tratarErroCodigo(w, err)
		return
	}

	responderJSON(w, http.StatusCreated, MensagemResponse{
		Mensagem: "Código vinculado ao salão com sucesso",
	})
}

// ------------------------------------------------------------
// ALTERAR CÓDIGO
// ------------------------------------------------------------

func (h *Handler) AlterarCodigo(
	w http.ResponseWriter,
	r *http.Request,
) {

	salaoID, err := obterSalaoID(r)
	if err != nil {
		responderErro(
			w,
			http.StatusBadRequest,
			"ID do salão inválido",
		)
		return
	}

	var request CodigoRequest

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		responderErro(
			w,
			http.StatusBadRequest,
			"Dados inválidos",
		)
		return
	}

	codigo := strings.TrimSpace(request.Codigo)

	err = h.controller.AlterarCodigo(salaoID, codigo)

	if err != nil {
		h.tratarErroCodigo(w, err)
		return
	}

	responderJSON(w, http.StatusOK, MensagemResponse{
		Mensagem: "Código do salão alterado com sucesso",
	})
}

// ------------------------------------------------------------
// TRATAMENTO DE ERROS
// ------------------------------------------------------------

func (h *Handler) tratarErroCodigo(
	w http.ResponseWriter,
	err error,
) {

	switch {
	case errors.Is(err, ErrSalaoNaoEncontrado):
		responderErro(
			w,
			http.StatusNotFound,
			"Salão não encontrado",
		 )

	case errors.Is(err, ErrCodigoDuplicado):
		responderErro(
			w,
			http.StatusConflict,
			"Código já existente. Troque o código.",
		 )

	case errors.Is(err, ErrCodigoInvalido):
		responderErro(
			w,
			http.StatusBadRequest,
			"Informe um código válido",
		)

	case errors.Is(err, ErrSalaoJaPossuiCodigo):
		responderErro(
			w,
			http.StatusConflict,
			"Este salão já possui um código de vínculo",
		)

	case errors.Is(err, ErrCodigoNaoCadastrado):
		responderErro(
			w,
			http.StatusConflict,
			"Este salão ainda não possui código para alteração",
		)

	default:
		responderErro(
			w,
			http.StatusInternalServerError,
			"Não foi possível realizar a operação",
		)
	}
}

// ------------------------------------------------------------
// OBTER ID DO SALÃO
// ------------------------------------------------------------

func obterSalaoID(
	r *http.Request,
) (uint64, error) {

	id := r.PathValue("salao_id")

	if id == "" {
		return 0, errors.New("ID do salão não informado")
	}

	salaoID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, err
	}

	if salaoID == 0 {
		return 0, errors.New("ID do salão inválido")
	}

	return salaoID, nil
}

// ------------------------------------------------------------
// RESPOSTA JSON
// ------------------------------------------------------------

func responderJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

// ------------------------------------------------------------
// RESPOSTA DE ERRO
// ------------------------------------------------------------

func responderErro(
	w http.ResponseWriter,
	status int,
	mensagem string,
) {

	responderJSON(w, status, ErrorResponse{
		Erro: mensagem,
	})
}