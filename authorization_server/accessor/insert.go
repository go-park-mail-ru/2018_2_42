// Годная книжка для тонкостей "database/sql".
// https://itjumpstart.files.wordpress.com/2015/03/database-driven-apps-with-go.pdf

package accessor

import (
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
)

// этот паттерн называется proxy - обёртка вокруг пучка соединений к базе данных
// не позволяет делать запросы в обход логики
type DB struct {
	*sql.DB // Резерв соединений к базе данных.
}

var Db DB // собственный тип, что бы прикреплять к нему функции с бизнес логикой.

func init() {
	newDb, err := sql.Open("postgres",
		"postgres://postgres:@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	} else {
		Db = DB{newDb}
	}
	// TODO: db.Close() в конце main.
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
		panic(err)
	}
}

// Коллекция идентификаторов предкомпилированных sql запросов.
// https://postgrespro.ru/docs/postgrespro/10/sql-prepare
var preparedStatements = map[string]*sql.Stmt{}

// По аналогии с regexp.MustCompile используется при запуске,
// проверяет успешность подготовки SQL для дальнейшего использования.
func must(stmt *sql.Stmt, err error) (*sql.Stmt) {
	if err != nil {
		panic(errors.New("error when compiling SQL expression: " + err.Error()))
	}
	return stmt
}

// Далее функции, реализующие логику зпросов к базе.

// В подобных init подготавливаются все зпросы SQL.
func init() {
	preparedStatements["insertIntoUser"] = must(Db.Prepare(`
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
	err = preparedStatements["insertIntoUser"].QueryRow(login, avatarAddress, disposable).Scan(&id)
	if err != nil {
		err = errors.New("Error on exec 'insertIntoUser' statment: " + err.Error())
	}
	return id, err
}

func init() {
	preparedStatements["insertIntoRegularLoginInformation"] = must(Db.Prepare(`
insert into "regular_login_information"(
	"user_id",
	"password_hash"
) values (
	$1, $2
);
	`))
}

func (db *DB) InsertIntoRegularLoginInformation(userId UserId, passwordHash string) (err error) {
	_, err = preparedStatements["insertIntoRegularLoginInformation"].Exec(userId, passwordHash)
	if err != nil {
		err = errors.New("Error on exec 'InsertIntoRegularLoginInformation' statment: " + err.Error())
	}
	return err
}

func init() {
	preparedStatements["insertIntoGameStatistics"] = must(Db.Prepare(`
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
	_, err = preparedStatements["insertIntoGameStatistics"].Exec(userId, gamesPlayed, wins)
	if err != nil {
		err = errors.New("Error on exec 'insertIntoGameStatistics' statment: " + err.Error())
	}
	return err
}

func init() {
	preparedStatements["insertIntoCurrentLogin"] = must(Db.Prepare(`
insert into "current_login" (
	"user_id",
    "authorization_token" -- cookie пользователя
) values (
	$1, $2
);
    `))
}

func (db *DB) InsertIntoCurrentLogin(userId UserId, authorizationToken string) (err error) {
	_, err = preparedStatements["insertIntoCurrentLogin"].Exec(userId, authorizationToken)
	if err != nil {
		err = errors.New("Error on exec 'InsertIntoGameStatistics' statment: " + err.Error())
	}
	return err
}

func init() {
	preparedStatements["selectLeaderBoard"] = must(Db.Prepare(`
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
			err = errors.New("Error on exec 'SelectLeaderBoard' statment: " + err.Error())
		}
	}()
	rows, err := preparedStatements["selectLeaderBoard"].Query(limit, offset)
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

func init() {
	preparedStatements["selectUserByLogin"] = must(Db.Prepare(`
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
	if err = preparedStatements["selectUserByLogin"].QueryRow(login).Scan(
		&userInformation.Login,
		&userInformation.AvatarAddress,
		&userInformation.GamesPlayed,
		&userInformation.Wins,
	); err != nil {
		err = errors.New("Error on exec 'SelectUserByLogin' statment: " + err.Error())
	}
	return
}
