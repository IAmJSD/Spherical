package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/jakemakesstuff/spherical/db"
)

func mapConfig() map[string]reflect.Value {
	m := map[string]reflect.Value{}
	v := reflect.Indirect(reflect.ValueOf(&configVar))
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		typeField := t.Field(i)
		objField := v.Field(i).Addr()
		m[typeField.Tag.Get("config")] = objField
	}
	return m
}

func addToConfig(key string, value json.RawMessage) {
	configVarLock.Lock()
	defer configVarLock.Unlock()
	c := mapConfig()
	val, ok := c[key]
	if !ok {
		if key == "setup_key" {
			// Do not log this. We know.
			return
		}
		_, _ = fmt.Fprintln(os.Stderr, "[config] The key", key, "is not in the configuration structure - ignoring!")
		return
	}
	err := json.Unmarshal(value, val.Interface())
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "[config] Failed to JSON unmarshal key", key, "-", err)
	}
}

// Watch is used to inspect changes to the configuration and initially set it up.
func Watch() error {
	err := db.InternallyConsumeConfig(addToConfig)
	if err == nil {
		fmt.Println("[config] Successfully started watching config!")
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "[config] Failed to get config.")
	}
	return err
}
