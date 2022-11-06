package public

// Just a shim to expose the public folder.

import (
	"embed"
	"fmt"
	"net/http"
	"os"
)

//go:embed * **/*
var public embed.FS

func GetFS(dev bool) http.FileSystem {
	if dev {
		s, err := os.Stat("public")
		if err == nil && s.IsDir() {
			return http.FS(os.DirFS("public"))
		}
		fmt.Println("[public] public folder is not in cwd - live updates are off!")
	}
	return http.FS(public)
}
