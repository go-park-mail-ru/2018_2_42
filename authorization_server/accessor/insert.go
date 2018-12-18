// Годная книжка для тонкостей "database/sql".
// https://itjumpstart.files.wordpress.com/2015/03/database-driven-apps-with-go.pdf

package accessor

import (
	"database/sql"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"strconv"

	"github.com/go-park-mail-ru/2018_2_42/authorization_server/types"
)

// собственный тип, что бы прикреплять к нему функции с бизнес логикой.
// паттерн - обёртка вокруг пула соединений к базе данных
type DB struct {
	*sql.DB // Резерв соединений к базе данных.
}

func ConnectToDatabase(DataSourceName string) (db DB, err error) {
	newDb, err := sql.Open("postgres", DataSourceName)
	if err != nil {
		err = errors.Wrap(err, "error on open connection to '"+DataSourceName+"'")
		return
	}

	err = newDb.Ping()
	if err != nil {
		err = errors.Wrap(err, "error during the first connection to '"+DataSourceName+"' (Are you sure that the database exists and the application has access to it?): ")
		return
	}

	db = DB{newDb}
	return
}

// подготовит все prepared statement,
// должна быть вызвана после соединения с базой.
func (db *DB) InitDatabase() (err error) {
	initAll := []func() error{
		db.init00,
		db.init01,
		db.init02,
		db.init03,
		db.init04,
		db.init05,
		db.init06,
		db.init07,
		db.init08,
		db.init09,
		db.init10,
	}
	for i, init := range initAll {
		err = init()
		if err != nil {
			err = errors.Wrap(err, "during preparing function 'accessor.init" + strconv.Itoa(i) + "': ")
			break
		}
	}
	return
}

// В подобных init подготавливаются все зпросы SQL.
var stmtInsertIntoUser *sql.Stmt

func (db *DB) init01() (err error) {
	//language=PostgreSQL
	stmtInsertIntoUser, err = db.Prepare(`
insert into "user"(
	"login",
	"avatar_address",
	"disposable",
	"last_login_time"
) values (
	$1, $2, $3, now()
) returning id as new_user_id;
	`)
	err = errors.Wrap(err, "init01: ")
	return
}

// Такие функции скрывают нетипизированность prepared statement.
func (db *DB) InsertIntoUser(login string, avatarAddress string, disposable bool) (id UserID, isDuplicate bool, err error) {
	err = stmtInsertIntoUser.QueryRow(login, avatarAddress, disposable).Scan(&id)
	if err != nil {
		isDuplicate = err.(*pq.Error).Code == "23505"
		if isDuplicate {
			err = nil
			return
		}
		err = errors.New("Error on exec 'insertIntoUser' statement: " + err.Error())
	}
	return
}

var stmtInsertIntoRegularLoginInformation *sql.Stmt

func (db *DB) init02() (err error) {
	//language=PostgreSQL
	stmtInsertIntoRegularLoginInformation, err = db.Prepare(`
insert into "regular_login_information"(
	"user_id",
	"password_hash"
) values (
	$1, $2
);
	`)
	err = errors.Wrap(err, "init02: ")
	return
}

func (db *DB) InsertIntoRegularLoginInformation(userID UserID, passwordHash string) (err error) {
	_, err = stmtInsertIntoRegularLoginInformation.Exec(userID, passwordHash)
	if err != nil {
		err = errors.New("Error on exec 'InsertIntoRegularLoginInformation' statement: " + err.Error())
	}
	return err
}

var stmtInsertIntoGameStatistics *sql.Stmt

func (db *DB) init03() (err error) {
	//language=PostgreSQL
	stmtInsertIntoGameStatistics, err = db.Prepare(`
insert into "game_statistics" (
	"user_id",
	"games_played",
	"wins"
) values (
	$1, $2, $3 
);
	`)
	err = errors.Wrap(err, "init03: ")
	return
}

func (db *DB) InsertIntoGameStatistics(userId UserID, gamesPlayed int32, wins int32) (err error) {
	_, err = stmtInsertIntoGameStatistics.Exec(userId, gamesPlayed, wins)
	if err != nil {
		err = errors.New("Error on exec 'insertIntoGameStatistics' statement: " + err.Error())
	}
	return err
}

var stmtInsertIntoCurrentLogin *sql.Stmt

func (db *DB) init04() (err error) {
	//language=PostgreSQL
	stmtInsertIntoCurrentLogin, err = db.Prepare(`
insert into "current_login" (
	"user_id",
    "authorization_token" -- cookie пользователя
) values (
	$1, $2
) on conflict ("user_id") do update set 
    "authorization_token" = excluded."authorization_token"
;   `)
	err = errors.Wrap(err, "init04: ")
	return
}

// update or insert
func (db *DB) UpsertIntoCurrentLogin(userID UserID, authorizationToken string) (err error) {
	_, err = stmtInsertIntoCurrentLogin.Exec(userID, authorizationToken)
	if err != nil {
		err = errors.New("Error on exec 'InsertIntoGameStatistics' statement: " + err.Error())
	}
	return err
}

var stmtSelectLeaderBoard *sql.Stmt

func (db *DB) init05() (err error) {
	//language=PostgreSQL
	stmtSelectLeaderBoard, err = db.Prepare(`
select
    "user"."login",
    "user"."avatar_address",
    "game_statistics"."games_played" as "gamesPlayed",
    "game_statistics"."wins"
from 
	"user",
	"game_statistics"
where 
	"user"."id" = "game_statistics".user_id
order by 
	"game_statistics"."wins" desc,
	"gamesPlayed" asc
limit
    $1
offset
    $2
;   `)
	err = errors.Wrap(err, "init05: ")
	return
}

func (db *DB) SelectLeaderBoard(limit int, offset int) (usersInformation types.PublicUsersInformation, err error) {
	defer func() {
		if err != nil {
			err = errors.New("Error on exec 'SelectLeaderBoard' statement: " + err.Error())
		}
	}()
	rows, err := stmtSelectLeaderBoard.Query(limit, offset)
	if err != nil {
		return
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return
		}
		userInformation := types.PublicUserInformation{}
		if err = rows.Scan(
			&userInformation.Login,
			&userInformation.AvatarAddress,
			&userInformation.GamesPlayed,
			&userInformation.Wins,
		); err != nil {
			return
		}
		usersInformation = append(usersInformation, userInformation)
	}
	return
}

var stmtSelectUserByLogin *sql.Stmt

func (db *DB) init06() (err error) {
	//language=PostgreSQL
	stmtSelectUserByLogin, err = db.Prepare(`
select
    "user"."login",
    "user"."avatar_address",
    "game_statistics"."games_played",
    "game_statistics"."wins"
from 
	"user",
	"game_statistics"
where 
	"user"."login" = $1 and
	"user"."id" = "game_statistics"."user_id"
;   `)
	err = errors.Wrap(err, "init03: ")
	return
}

func (db *DB) SelectUserByLogin(login string) (userInformation types.PublicUserInformation, err error) {
	if err = stmtSelectUserByLogin.QueryRow(login).Scan(
		&userInformation.Login,
		&userInformation.AvatarAddress,
		&userInformation.GamesPlayed,
		&userInformation.Wins,
	); err != nil {
		err = errors.New("Error on exec 'SelectUserByLogin' statement: " + err.Error())
	}
	return
}

var stmtSelectUserIDByLoginPassword *sql.Stmt

func (db *DB) init07() (err error) {
	//language=PostgreSQL
	stmtSelectUserIDByLoginPassword, err = db.Prepare(`
select
	"user"."id"
from 
	"user",
	"regular_login_information"
where 
	"user"."login" = $1 and
	"user"."id" = "regular_login_information"."user_id" and 
	"regular_login_information"."password_hash" = $2
;   `)
	err = errors.Wrap(err, "init07: ")
	return
}

func (db *DB) SelectUserIdByLoginPasswordHash(login string, passwordHash string) (exist bool, userId UserID, err error) {
	err = stmtSelectUserIDByLoginPassword.QueryRow(login, passwordHash).Scan(
		&userId,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			err = nil
			// exist == false as default.
		} else {
			err = errors.New("Error on exec 'SelectUserIdByLoginPasswordHash' statement: " + err.Error())
		}
	} else {
		exist = true
	}
	return
}

var stmtDropUsersSession *sql.Stmt

func (db *DB) init08() (err error) {
	//language=PostgreSQL
	stmtDropUsersSession, err = db.Prepare(`
delete from
    "current_login"
where 
    "current_login"."authorization_token" = $1
;   `)
	err = errors.Wrap(err, "init08: ")
	return
}

func (db *DB) DropUsersSession(authorizationToken string) (err error) {
	_, err = stmtDropUsersSession.Exec(authorizationToken)
	if err != nil {
		err = errors.New("Error on exec 'dropUsersSession' statement: " + err.Error())
	}
	return err
}

var stmtUpdateUsersAvatarByLogin *sql.Stmt

func (db *DB) init09() (err error) {
	//language=PostgreSQL
	stmtUpdateUsersAvatarByLogin, err = db.Prepare(`
update
    "user"
set
    "avatar_address" = $2
from
    "current_login"
where
    "user"."login" = $1
;    `)
	err = errors.Wrap(err, "init09: ")
	return
}

func (db *DB) UpdateUsersAvatarByLogin(login string, avatarAddress string) (err error) {
	result, err := stmtUpdateUsersAvatarByLogin.Exec(login, avatarAddress)
	if err != nil {
		err = errors.New("Error on exec 'UpdateUsersAvatarByLogin' statement: " + err.Error())
		return
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		err = errors.New("user unknown")
	}
	return
}

var stmtSelectUserLoginBySessionID *sql.Stmt

func (db *DB) init10() (err error) {
	//language=PostgreSQL
	stmtSelectUserLoginBySessionID, err = db.Prepare(`
select 	
	"user"."id",
	"user"."login",
	"user"."avatar_address",
	"user"."disposable",
	"user"."last_login_time"
from 
	"user", "current_login"
where
	"current_login"."authorization_token" = $1 and
	"current_login"."user_id" = "user"."id"
;    `)
	err = errors.Wrap(err, "init10: ")
	return
}

func (db *DB) SelectUserBySessionId(authorizationToken string) (exist bool, user User, err error) {
	err = stmtSelectUserLoginBySessionID.QueryRow(authorizationToken).Scan(
		&user.Id,
		&user.Login,
		&user.AvatarAddress,
		&user.Disposable,
		&user.LastLoginTime,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			err = nil
			// exist == false as default.
		} else {
			err = errors.New("Error on exec 'SelectUserBySid' statment: " + err.Error())
		}
	} else {
		exist = true
	}
	return
}
