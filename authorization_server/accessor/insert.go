// Годная книжка для тонкостей "database/sql".
// https://itjumpstart.files.wordpress.com/2015/03/database-driven-apps-with-go.pdf

package accessor

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"reflect"

	"github.com/go-park-mail-ru/2018_2_42/authorization_server/config"
)

// этот паттерн называется proxy - обёртка вокруг пучка соединений к базе данных
// не позволяет делать запросы в обход логики
type DB struct {
	*sql.DB // Резерв соединений к базе данных.
}

var Db DB // собственный тип, что бы прикреплять к нему функции с бизнес логикой.

func init() {
	dataSourceName := ""
	if data, err := config.ParseConfig(); err == nil {
		if reflect.TypeOf(data["data_source_name"]).Kind() == reflect.String {
			dataSourceName = data["data_source_name"].(string)
		} else {
			log.Fatal(errors.New(fmt.Sprintf(
				"not string 'data_source_name' configuration param '%v'", data["data_source_name"])))
		}
	} else {
		log.Fatal(err)
	}

	newDb, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(err)
	} else {
		Db = DB{newDb}
	}

	err = Db.Ping()
	if err != nil {
		log.Fatal(errors.New("error during the first connection (Are you sure that the database exists and the application has access to it?): " + err.Error()))
	}

	_, err = Db.Exec(`
begin transaction;

create table if not exists "user" (
  "id"              serial4 primary key,
  -- видимое другим игрокам имя пользователя
  "login"           text      not null unique,
  -- адрес относительно корня сайта: '/media/name-src32.ext'
  "avatar_address"  text      not null,
  -- Если True - пользователь не залогинен, играет просто так.
  -- Такие пользователи создаются, когда входят в игру с одним только именем,
  -- и удаляются при выходе из партии.
  "disposable"      boolean   not null,
  "last_login_time" timestamp not null
);

-- не более одной строчки на пользователя в нижних трёх таблицах
create table if not exists "regular_login_information" (
  "id"            serial4 primary key,
  "user_id"       integer not null unique references "user" ("id") on delete cascade,
  "password_hash" text not null
);

-- данные для таблицы лидеров
create table if not exists "game_statistics" (
  "id"           serial4 primary key,
  "user_id"      integer not null unique references "user" ("id") on delete cascade,
  "games_played" integer not null, -- количество начатых игр
  "wins"         integer not null -- количество доведённых до победного конца
);

-- текущая принадлежность к игре.
-- допущение - только одна игра в один момент времени.
-- Сама игра не отображается никак в базе, только результаты.
create table if not exists "current_login" (
  "id"                  serial4 primary key,
  "user_id"             integer not null unique references "user" ("id") on delete cascade,
  -- токен авторицации, ставящийся как cookie пользователю
  "authorization_token" text null unique
);

commit;
	`)
	if err != nil {
		log.Fatal(errors.New("error during preparation database tables: " + err.Error()))
	}
}

// По аналогии с regexp.MustCompile используется при запуске,
// проверяет успешность подготовки SQL для дальнейшего использования.
func must(stmt *sql.Stmt, err error) *sql.Stmt {
	if err != nil {
		log.Fatal(errors.New("error when compiling SQL expression: " + err.Error()))
	}
	return stmt
}

// Далее функции, реализующие логику зпросов к базе.

// В подобных init подготавливаются все зпросы SQL.
var stmtInsertIntoUser *sql.Stmt

func init() {
	stmtInsertIntoUser = must(Db.Prepare(`
insert into "user"(
	"login",
	"avatar_address",
	"disposable",
	"last_login_time"
) values (
	$1, $2, $3, now()
) returning id as new_user_id;
	`))
}

// Такие функции скрывают нетипизированность prepared statement.
func (db *DB) InsertIntoUser(login string, avatarAddress string, disposable bool) (id UserId, err error) {
	err = stmtInsertIntoUser.QueryRow(login, avatarAddress, disposable).Scan(&id)
	if err != nil {
		err = errors.New("Error on exec 'insertIntoUser' statement: " + err.Error())
	}
	return id, err
}

var stmtInsertIntoRegularLoginInformation *sql.Stmt

func init() {
	stmtInsertIntoRegularLoginInformation = must(Db.Prepare(`
insert into "regular_login_information"(
	"user_id",
	"password_hash"
) values (
	$1, $2
);
	`))
}

func (db *DB) InsertIntoRegularLoginInformation(userId UserId, passwordHash string) (err error) {
	_, err = stmtInsertIntoRegularLoginInformation.Exec(userId, passwordHash)
	if err != nil {
		err = errors.New("Error on exec 'InsertIntoRegularLoginInformation' statement: " + err.Error())
	}
	return err
}

var stmtInsertIntoGameStatistics *sql.Stmt

func init() {
	stmtInsertIntoGameStatistics = must(Db.Prepare(`
insert into "game_statistics" (
	"user_id",
	"games_played",
	"wins"
) values (
	$1, $2, $3 
);
	`))
}

func (db *DB) InsertIntoGameStatistics(userId UserId, gamesPlayed int32, wins int32) (err error) {
	_, err = stmtInsertIntoGameStatistics.Exec(userId, gamesPlayed, wins)
	if err != nil {
		err = errors.New("Error on exec 'insertIntoGameStatistics' statement: " + err.Error())
	}
	return err
}

var stmtInsertIntoCurrentLogin *sql.Stmt

func init() {
	stmtInsertIntoCurrentLogin = must(Db.Prepare(`
insert into "current_login" (
	"user_id",
    "authorization_token" -- cookie пользователя
) values (
	$1, $2
) on conflict ("user_id") do update set 
    "authorization_token" = excluded."authorization_token"
;   `))
}

// update or insert
func (db *DB) UpsertIntoCurrentLogin(userId UserId, authorizationToken string) (err error) {
	_, err = stmtInsertIntoCurrentLogin.Exec(userId, authorizationToken)
	if err != nil {
		err = errors.New("Error on exec 'InsertIntoGameStatistics' statement: " + err.Error())
	}
	return err
}

var stmtSelectLeaderBoard *sql.Stmt

func init() {
	stmtSelectLeaderBoard = must(Db.Prepare(`
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
;   `))
}

func (db *DB) SelectLeaderBoard(limit int, offset int) (usersInformation []PublicUserInformation, err error) {
	defer func() {
		if err != nil {
			err = errors.New("Error on exec 'SelectLeaderBoard' statement: " + err.Error())
		}
	}()
	rows, err := stmtSelectLeaderBoard.Query(limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return
		}
		userInformation := PublicUserInformation{}
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

func init() {
	stmtSelectUserByLogin = must(Db.Prepare(`
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
;   `))
}

func (db *DB) SelectUserByLogin(login string) (userInformation PublicUserInformation, err error) {
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

var stmtSelectUserIdByLoginPassword *sql.Stmt

func init() {
	stmtSelectUserIdByLoginPassword = must(Db.Prepare(`
select
	"user"."id"
from 
	"user",
	"regular_login_information"
where 
	"user"."login" = $1 and
	"user"."id" = "regular_login_information"."user_id" and 
	"regular_login_information"."password_hash" = $2
;   `))
}

func (db *DB) SelectUserIdByLoginPasswordHash(login string, passwordHash string) (exist bool, userId UserId, err error) {
	err = stmtSelectUserIdByLoginPassword.QueryRow(login, passwordHash).Scan(
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

func init() {
	stmtDropUsersSession = must(Db.Prepare(`
delete from
    "current_login"
where 
    "current_login"."authorization_token" = $1
;   `))
}

func (db *DB) DropUsersSession(authorizationToken string) (err error) {
	_, err = stmtDropUsersSession.Exec(authorizationToken)
	if err != nil {
		err = errors.New("Error on exec 'dropUsersSession' statement: " + err.Error())
	}
	return err
}

var stmtUpdateUsersAvatarByLogin *sql.Stmt

func init() {
	stmtUpdateUsersAvatarByLogin = must(Db.Prepare(`
update
    "user"
set
    "avatar_address" = $2
from
    "current_login"
where
    "user"."login" = $1
;    `))
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

var stmtSelectUserLoginBySessionId *sql.Stmt

func init() {
	stmtSelectUserLoginBySessionId = must(Db.Prepare(`
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
;    `))
}

func (db *DB) SelectUserBySessionId(authorizationToken string) (exist bool, user User, err error) {
	err = stmtSelectUserLoginBySessionId.QueryRow(authorizationToken).Scan(
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
