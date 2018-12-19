// Схема базы данных, c которой работает сервис.
// Скомпилированный бинарник содержит схему, это серьёзно упрощает первое развёртывание.

package accessor

import "github.com/pkg/errors"

func (db *DB) init00() (err error) {
	//language=PostgreSQL
	_, err = db.Exec(`
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
	err = errors.Wrap(err, "error during preparation database tables: init00: ")
	return
}
