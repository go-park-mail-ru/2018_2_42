package database

import (
	"auth/models"
	"fmt"
	"log"
)

func SelectUserByLogin(login string) (userProfile models.UserInfo, err error) {
	tx := StartTransaction()
	defer tx.Rollback()

	rows := tx.QueryRow(` 
		SELECT u."login", u."avatar_address", g."games_played", g."wins"
		FROM users u
		JOIN game_statistics g
		ON u."id" = g."user_id" AND u."login" = $1`,
		&login)

	err = rows.Scan(&userProfile.Login, &userProfile.AvatarAddress, &userProfile.GamesPlayed, &userProfile.Wins)
	if err != nil {
		log.Println(err)
		return userProfile, err
	}

	fmt.Println("here")
	CommitTransaction(tx)
	return
}
