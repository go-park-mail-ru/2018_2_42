package types

// ответ сервера на всякие действия.
type ServerResponse struct {
	Status string `json:"status"`
	Message string `json:"message"`
}

type NewUserRegistration struct {
	Login string `json:"login"`
	Password string `json:"password"`
}
