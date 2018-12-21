package environment

import (
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/AWS_upload_awatar"
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/accessor"
)

type Environment struct {
	DB     accessor.DB
	Config Config
	AWSUploader *AWS_upload_awatar.AWSUploader
}

type Config struct {
	PostgresPath  *string
	ListeningPort *string
	ImagesRoot    *string
}
