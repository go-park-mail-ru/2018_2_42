package environment

import (
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/accessor"
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/config"
)

type Environment struct {
	DB     accessor.DB
	Config config.Config
}
