package comments

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Listar(w http.ResponseWriter, r *http.Request) {
	var req ListRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "JSON inválido",
		})
		return
	}

	data, err := h.service.Listar(r.Context(), req.IDSalao)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"comentarios": data,
		},
	})
}

func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	filePath, cleanup, err := receberFormulario(w, r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	defer cleanup()

	idSalao, _ := strconv.ParseInt(r.FormValue("id_salao"), 10, 64)

	req := CreateRequest{
		IDSalao:    idSalao,
		Nome:       r.FormValue("nome"),
		Comentario: r.FormValue("comentario"),
	}

	if err := h.service.Criar(r.Context(), req, filePath); err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Comentário criado com sucesso!",
	})
}

func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	filePath, cleanup, err := receberFormulario(w, r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	defer cleanup()

	idSalao, _ := strconv.ParseInt(r.FormValue("id_salao"), 10, 64)
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)

	ativoStr := strings.ToLower(strings.TrimSpace(r.FormValue("ativo")))
	ativo := ativoStr == "on" || ativoStr == "true" || ativoStr == "1"

	req := UpdateRequest{
		IDSalao:    idSalao,
		ID:         id,
		Nome:       r.FormValue("nome"),
		Comentario: r.FormValue("comentario"),
		Ativo:      ativo,
	}

	if err := h.service.Atualizar(r.Context(), req, filePath); err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Comentário atualizado com sucesso!",
	})
}

func (h *Handler) Excluir(w http.ResponseWriter, r *http.Request) {
	var req DeleteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "JSON inválido",
		})
		return
	}

	if err := h.service.Excluir(r.Context(), req); err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Comentário removido com sucesso!",
	})
}

func receberFormulario(
	w http.ResponseWriter,
	r *http.Request,
) (string, func(), error) {
	r.Body = http.MaxBytesReader(w, r.Body, 6*1024*1024)

	if err := r.ParseMultipartForm(5 * 1024 * 1024); err != nil {
		return "", func() {}, errors.New("formulário inválido ou muito grande")
	}

	var cleanupFiles []string

	cleanup := func() {
		for _, path := range cleanupFiles {
			_ = os.Remove(path)
		}
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}

	file, header, err := r.FormFile("img")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", cleanup, nil
		}
		return "", cleanup, errors.New("não foi possível receber a imagem")
	}
	defer file.Close()

	if header.Size > 5*1024*1024 {
		return "", cleanup, errors.New("a imagem deve ter no máximo 5 MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	if !allowed[ext] {
		return "", cleanup, errors.New("use uma imagem JPG, PNG ou WEBP")
	}

	content, err := io.ReadAll(io.LimitReader(file, 5*1024*1024+1))
	if err != nil || len(content) > 5*1024*1024 {
		return "", cleanup, errors.New("não foi possível ler a imagem ou ela excede 5 MB")
	}

	tempDir := filepath.Join(os.TempDir(), "salao-comments")
	if err := os.MkdirAll(tempDir, 0700); err != nil {
		return "", cleanup, errors.New("não foi possível preparar o arquivo")
	}

	temp, err := os.CreateTemp(tempDir, "comment-*"+ext)
	if err != nil {
		return "", cleanup, errors.New("não foi possível salvar a imagem temporária")
	}

	if _, err := temp.Write(content); err != nil {
		temp.Close()
		os.Remove(temp.Name())
		return "", cleanup, errors.New("não foi possível salvar a imagem")
	}

	if err := temp.Close(); err != nil {
		os.Remove(temp.Name())
		return "", cleanup, errors.New("não foi possível finalizar a imagem")
	}

	cleanupFiles = append(cleanupFiles, temp.Name())
	return temp.Name(), cleanup, nil
}

func handleError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "Erro ao processar comentários"

	switch {
	case errors.Is(err, ErrInvalidData):
		status = http.StatusBadRequest
		message = "Informe os dados obrigatórios corretamente"
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
		message = "Comentário não encontrado"
	}

	writeJSON(w, status, map[string]any{
		"success": false,
		"message": message,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
