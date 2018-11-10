package database

import (
	"auth/models"
)

func InsertIntoUser(login, password, avatar string, disposable bool) (id models.UserID, err error) {
	tx := StartTransaction()
	defer tx.Rollback()

	rows := tx.QueryRow(` 
		INSERT INTO "user" ("login", "password", "avatar_address", "disposable","last_login_time") 
		VALUES ($1, $2, $3, now()) 
		RETURNING id`, &login, &password, &avatar, &disposable)

	err = rows.Scan(&id)
	if err != nil {
		return id, err
	}

	return id, nil
}
