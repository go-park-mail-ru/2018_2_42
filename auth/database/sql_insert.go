package database

import (
	"auth/models"
)

func InsertIntoUser(login, password, avatar string, disposable bool) (id models.UserID, err error) {
	tx := StartTransaction()
	defer tx.Rollback()

	rows := tx.QueryRow(` 
		INSERT INTO "users" ("login", "password_hash", "avatar_address", "last_login_time") 
		VALUES ($1, $2, $3, now()) 
		RETURNING id`,
		&login, &password, &avatar)

	err = rows.Scan(&id)
	if err != nil {
		return id, err
	}

	CommitTransaction(tx)
	return id, nil
}

func InsertIntoGameStatistics(userID models.UserID, gamesPlayed, wins int) (err error) {
	tx := StartTransaction()
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO "game_statistics" ("user_id", "games_played", "wins") 
		VALUES ($1, $2, $3)`,
		&userID, &gamesPlayed, &wins)

	if err != nil {
		return err
	}

	CommitTransaction(tx)
	return nil
}

func InsertIntoCurrentLogin(userID models.UserID, authorizationToken string) (err error) {
	tx := StartTransaction()
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO "current_login" ("user_id", "authorization_token") 
		VALUES ($1, $2) 
		ON CONFLICT ("user_id") 
		DO UPDATE SET "authorization_token" = excluded."authorization_token"`,
		&userID, &authorizationToken)

	if err != nil {
		return err
	}

	CommitTransaction(tx)
	return nil
}
