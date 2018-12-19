package types

// общая форма ответа сервера.

//easyjson:json
type ServerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

//easyjson:json
type NewUserRegistration struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Публичная информация пользователя
//easyjson:json
type PublicUserInformation struct {
	Login         string `json:"login"`
	AvatarAddress string `json:"avatarAddress"`
	GamesPlayed   int    `json:"gamesPlayed"`
	Wins          int    `json:"wins"`
}

//easyjson:json
type PublicUsersInformation []PublicUserInformation
