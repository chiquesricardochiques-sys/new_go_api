package profissional

type AddRelationsRequest struct {
	IDSalao    int64  `json:"id_salao"`
	AuthUserID int64  `json:"auth_user_id"`
	Nome       string `json:"nome"`
}

type VerifyRequest struct {
	IDSalao    int64 `json:"id_salao"`
	AuthUserID int64 `json:"auth_user_id"`
}

type VerifyResponse struct {
	Existe    bool   `json:"existe"`
	Ativo     bool   `json:"ativo"`
	IDSalao   int64  `json:"id_salao"`
	AuthUserID int64 `json:"auth_user_id"`
	Status    string `json:"status,omitempty"`
}