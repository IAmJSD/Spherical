package db

import (
	"net"

	"github.com/bwmarrin/snowflake"
	farmhash "github.com/leemcloughlin/gofarmhash"
)

func getDirtyNodeId() uint64 {
	ifas, _ := net.Interfaces()
	x := ""
	for _, ifa := range ifas {
		x += ifa.HardwareAddr.String()
	}
	return farmhash.Hash64([]byte(x))
}

var snowGen *snowflake.Node

func init() {
	var err error
	snowGen, err = snowflake.NewNode(int64(getDirtyNodeId() % 1023))
	if err != nil {
		panic(err)
	}
}
