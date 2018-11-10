begin transaction;

create table if not exists "user" (
  "id"              serial4 primary key,
  "login"           text      not null unique,
  "password_hash"   text not null,
  "avatar_address"  text      not null, -- адрес относительно корня сайта: '/media/name-src32.ext'
  "disposable"      boolean   not null,
  "last_login_time" timestamp not null
);

create table if not exists "game_statistics" (
  "id"           serial4 primary key,
  "user_id"      integer not null unique references "user" ("id") on delete cascade,
  "games_played" integer not null, -- количество начатых игр
  "wins"         integer not null -- количество доведённых до победного конца
);

create table if not exists "current_login" (
  "id"                  serial4 primary key,
  "user_id"             integer not null unique references "user" ("id") on delete cascade,
  -- токен авторицации, ставящийся как cookie пользователю
  "authorization_token" text null unique
);

commit;

