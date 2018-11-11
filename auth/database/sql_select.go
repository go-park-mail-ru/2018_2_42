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

	CommitTransaction(tx)
	return
}

func SelectLeaderBoard(limit, offset string) (*models.Users, error) {
	tx := StartTransaction()
	defer tx.Rollback()

	rows, err := tx.Query(` 
		SELECT u."login", u."avatar_address", g."games_played", g."wins"
		FROM users u
		JOIN game_statistics g
		ON u."id" = g."user_id"
		ORDER BY g."wins" DESC, g."games_played" ASC
		LIMIT $1
		OFFSET $2`,
		&limit, &offset)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	users := models.Users{}
	for rows.Next() {
		user := models.UserInfo{}

		if err = rows.Scan(
			&user.Login,
			&user.AvatarAddress,
			&user.GamesPlayed,
			&user.Wins,
		); err != nil {
			fmt.Println(err)
		}
		users = append(users, &user)
	}

	CommitTransaction(tx)
	return &users, nil
}
