package config

import (
	"github.com/pkg/errors"
	"github.com/yosuke-furukawa/json5/encoding/json5"
	"os"
	"sync"
)

var configOnce = struct {
	sync.Once
	data map[string]interface{}
	err  error
}{
	data: make(map[string]interface{}),
}

// Return config, can be called from anywhere.
func ParseConfig() (map[string]interface{}, error) {
	configOnce.Do(func() {
		var file *os.File
		file, configOnce.err = os.Open("./main.json5")
		if configOnce.err != nil {
			configOnce.err = errors.New("error on Open('./main.json5'): Do you put config next to the application? : " + configOnce.err.Error())
			return
		}
		defer file.Close()
		dec := json5.NewDecoder(file)
		configOnce.err = dec.Decode(&configOnce.data)
		if configOnce.err != nil {
			configOnce.err = errors.New("error on parsing config : it must be json object: " + configOnce.err.Error())
			return
		}
		return
	})
	return configOnce.data, configOnce.err
}
