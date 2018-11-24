package models

import (
	"database/sql"
	"time"
)

type UserID int32

//easyjson:json
type User struct {
	ID            UserID    `json:"id"`
	Login         string    `json:"login"`
	PasswordHash  string    `json:"password_hash"`
	AvatarAddress string    `json:"avatar_address"`
	LastLoginTime time.Time `json:"last_login_time"`
}

//easyjson:json
type UserRegistration struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

//easyjson:json
type UserGameStatistic struct {
	UserID      UserID `json:"user_id"`
	GamesPlayed int32  `json:"games_played"`
	Wins        int    `json:"wins"`
}

//easyjson:json
type CurrentLogin struct {
	UserID             UserID         `json:"user_id"`
	AuthorizationToken sql.NullString `json:"authorization_token"`
}

//easyjson:json
type UserInfo struct {
	Login         string `json:"login"`
	AvatarAddress string `json:"avatarAddress"`
	GamesPlayed   int    `json:"gamesPlayed"`
	Wins          int    `json:"wins"`
}

//easyjson:json
type Users []*UserInfo
