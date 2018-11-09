package models

import (
	"database/sql"
	"time"
)

type UserID int32

type User struct {
	ID            UserID    // primary key through which other fields are connected.
	Login         string    // visible to other players
	AvatarAddress string    // address relative to the root of the site: '/media/name-src32.ext'
	LastLoginTime time.Time // timestamp
	Disposable    bool      // temporary or not
}

//easyjson:json
type UserRegistration struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegularLoginInfo struct {
	UserID       UserID
	PasswordHash string // just SHA (TODO: sault)
}

type UserGameStatistic struct {
	UserID      UserID
	GamesPlayed int32 // count of games
	Wins        int   // count of winnings
}

type CurrentLogin struct {
	UserID             UserID
	AuthorizationToken sql.NullString //just coolkie
	CSRFToken          sql.NullString
}

type PublicUserInfo struct {
	Login         string `json:"login"`
	AvatarAddress string `json:"avatarAddress"`
	GamesPlayed   int    `json:"gamesPlayed"`
	Wins          int    `json:"wins"`
}

type ServerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
