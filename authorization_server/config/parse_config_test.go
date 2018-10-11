package config

import (
	"reflect"
	"testing"
)

func TestTimeConsuming(t *testing.T) {
	config, err := ParseConfig()
	if err != nil {
		t.Fatal(err)
	}
	if databasePath, ok := config["data_source_name"]; ok {
		if reflect.TypeOf(databasePath).Kind() == reflect.String {
			// It's fine.
		} else {
			t.Fatal("config['data_source_name'] must be string ( https://golang.org/pkg/database/sql/#Open )")
		}
	} else {
		t.Fatal("config must have field 'data_source_name' ( https://golang.org/pkg/database/sql/#Open )")
	}
}
