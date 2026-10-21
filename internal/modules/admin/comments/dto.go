package comments

type ListRequest struct {
	IDSalao int64 `json:"id_salao"`
}

type CreateRequest struct {
	IDSalao    int64  `json:"id_salao"`
	Nome       string `json:"nome"`
	Comentario string `json:"comentario"`
}

type UpdateRequest struct {
	IDSalao    int64  `json:"id_salao"`
	ID         int64  `json:"id"`
	Nome       string `json:"nome"`
	Comentario string `json:"comentario"`
	Ativo      bool   `json:"ativo"`
}

type DeleteRequest struct {
	IDSalao int64 `json:"id_salao"`
	ID      int64 `json:"id"`
}

type Comment struct {
	ID         int64  `json:"id"`
	Nome       string `json:"nome"`
	Comentario string `json:"comentario"`
	Img        string `json:"img"`
	Ativo      bool   `json:"ativo"`
}

type CommentImage struct {
	URL      string
	PublicID string
}
