package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jakemakesstuff/spherical/db"
)

func addToConfig(key string, value json.RawMessage) {
	switch key {
	case "hello":
		println(string(value))
	}
}

// Watch is used to inspect changes to the configuration and initially set it up.
func Watch() error {
	err := db.InternallyConsumeConfig(addToConfig)
	if err == nil {
		fmt.Println("[config] Successfully started watching config.")
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "[config] Failed to get config.")
	}
	return err
}
