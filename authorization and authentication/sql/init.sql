begin transaction;

create table if not exists "user" (
  "id"              serial4 primary key,
  -- видимое другим игрокам имя пользователя
  "login"           text      not null unique,
  -- адрес относительно корня сайта: '/media/name-src32.ext'
  "avatar_address"  text      not null unique,
  -- Если True - пользователь не залогинен, играет просто так.
  -- Такие пользователи создаются, когда входят в игру с одним только именем,
  -- и удаляются при выходе из партии.
  "disposable"      boolean   not null,
  "last_login_time" timestamp not null
);

-- не более одной строчки на пользователя в нижних трёх таблицах
create table if not exists "regular_login_information" (
  "id"            serial4 primary key,
  "user_id"       integer not null references "user" ("id") unique,
  "password_hash" text not null
);

-- данные для таблицы лидеров
create table if not exists "game_statistics" (
  "id"           serial4 primary key,
  "user_id"      integer not null references "user" ("id") unique,
  "games_played" integer not null, -- количество начатых игр
  "wins"         integer not null -- количество доведённых до победного конца
);

-- текущая принадлежность к игре.
-- допущение - только одна игра в один момент времени.
create table if not exists "current_login" (
  "id"                  serial4 primary key,
  "user_id"             integer not null references "user" ("id") unique,
  -- токен авторицации, ставящийся как cookie пользователю
  "authorization_token" text not null unique
);

commit;