package database

import "log"

func UpdateUsersAvatarByLogin(login, path string) error {
	tx := StartTransaction()
	defer tx.Rollback()

	result, err := tx.Exec(` 
	UPDATE "users"
	SET "avatar_address" = $2
	FROM "current_login"
	WHERE "userS"."login" = $1`,
		&login, &path)

	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		log.Println(err.Error())
		return err
	}

	CommitTransaction(tx)
	return nil
}
