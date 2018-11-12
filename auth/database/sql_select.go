package database

import (
	"auth/models"
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

func SelectUserBySession(token string) (userProfile models.User, err error) {
	tx := StartTransaction()
	defer tx.Rollback()

	rows := tx.QueryRow(` 
		SELECT u."login", u."password_hash", u."avatar_address", u."last_login_time", 
		FROM users u
		JOIN current_login c
		ON u."id" = c."user_id" AND c."authorization_token" = $1`,
		&token)

	err = rows.Scan(&userProfile.Login, &userProfile.PasswordHash, &userProfile.AvatarAddress, &userProfile.LastLoginTime)
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
			log.Println(err)
			return nil, err
		}
		users = append(users, &user)
	}

	CommitTransaction(tx)
	return &users, nil
}

func SelectUserIdByLoginPasswordHash(login, passwordHash string) (userID models.UserID, err error) {
	tx := StartTransaction()
	defer tx.Rollback()

	rows := tx.QueryRow(` 
		SELECT "id"
		FROM users
		WHERE "login" = $1 AND "password_hash" = $2`,
		&login, &passwordHash)

	err = rows.Scan(&userID)
	if err != nil {
		log.Println(err)
		return userID, err
	}

	CommitTransaction(tx)
	return
}

func DropUserSession(cookie string) error {
	tx := StartTransaction()
	defer tx.Rollback()

	query := `
		DELETE 
		FROM "current_login"
		WHERE "authorization_token" = $1`

	if _, err := tx.Exec(query, cookie); err != nil {
		log.Println(err)
		return err
	}

	CommitTransaction(tx)
	return nil
}
