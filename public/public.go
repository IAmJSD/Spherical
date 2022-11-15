package public

// Just a shim to expose the public folder.

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"sync/atomic"
)

//go:embed * **/*
var public embed.FS

var once uintptr

func GetFS(dev bool) fs.FS {
	if dev {
		s, err := os.Stat("public")
		if err == nil && s.IsDir() {
			if atomic.SwapUintptr(&once, 1) == 0 {
				fmt.Println("[public] dev mode enabled - using local fs for public folder!")
			}
			return os.DirFS("public")
		}
		if atomic.SwapUintptr(&once, 1) == 0 {
			fmt.Println("[public] public folder is not in cwd - live updates are off!")
		}
	}
	return public
}
