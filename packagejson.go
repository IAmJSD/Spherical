package packagejson

import (
	_ "embed"
	"encoding/json"
)

//go:embed package.json
var packageJson []byte

type PackageJson struct {
	Version string `json:"version"`
}

func Package() PackageJson {
	var x PackageJson
	err := json.Unmarshal(packageJson, &x)
	if err != nil {
		panic(err)
	}
	return x
}
