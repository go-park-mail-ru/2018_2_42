package environment

import (
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/accessor"
)

type Environment struct {
	DB     accessor.DB
	Config Config
}

type Config struct {
	PostgresPath  *string
	ListeningPort *string
	ImagesRoot    *string
}
