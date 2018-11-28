package config

import (
	"github.com/pkg/errors"
	"github.com/yosuke-furukawa/json5/encoding/json5"
	"os"
)

// json 5 используется для парсинга json с пропуском комментариев.
// Стандарт json это не поддерживает, а перезапись поля
// {"field": "comment", "field": "real value"} выглядит странно.

func ParseConfig(configPath string) (config Config, err error) {
	file, err := os.Open(configPath)
	if err != nil {
		err = errors.Wrap(err, "error on Open('"+configPath+"'): Do you put config next to the application? : ")
		return
	}
	dec := json5.NewDecoder(file)
	err = dec.Decode(&config)
	if err != nil {
		err = errors.Wrap(err, "on parsing config: it must be json string to string object: ")
		return
	}
	err = file.Close()
	if err != nil {
		err = errors.Wrap(err, "on close config: ")
		return
	}
	return
}
