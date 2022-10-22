package random

import (
	"crypto/rand"
	"encoding/binary"
	mathRand "math/rand"
)

type cryptoRandSource struct{}

func (cryptoRandSource) Int63() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int64(binary.LittleEndian.Uint64(b[:]) & (1<<63 - 1))
}

func (cryptoRandSource) Seed(int64) {}

func New() *mathRand.Rand {
	return mathRand.New(cryptoRandSource{})
}
