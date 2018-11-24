BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS "users" (
  "id"              SERIAL    PRIMARY KEY,
  "login"           TEXT      NOT NULL UNIQUE,
  "password_hash"   TEXT      NOT NULL,
  "avatar_address"  TEXT      NOT NULL, -- адрес относительно корня сайта: '/media/name-src32.ext'`
  "last_login_time" TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS "game_statistics" (
  "user_id"      INTEGER NOT NULL UNIQUE REFERENCES "users" ("id") ON DELETE CASCADE,
  "games_played" INTEGER NOT NULL, 
  "wins"         INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS "current_login" (
  "user_id"             INTEGER NOT NULL UNIQUE REFERENCES "users" ("id") ON DELETE CASCADE,
  "authorization_token" TEXT    NULL UNIQUE -- токен авторицации, ставящийся как cookie пользователю
);

COMMIT;

