package main

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jakemakesstuff/spherical/hashverifier"
)

type redisDriver struct {
	redisClient *redis.Client
}

func (r redisDriver) LookupPGP(hostname string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	hostname = strings.ToLower(hostname)
	b, err := r.redisClient.Get(ctx, "pgp:"+hostname).Bytes()
	if err != nil {
		return nil
	}
	return b
}

func (r redisDriver) WritePGP(hostname string, key []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	hostname = strings.ToLower(hostname)
	r.redisClient.Set(ctx, "pgp:"+hostname, key, time.Hour*24*30)
}

func (r redisDriver) Exists(hash []byte) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	hashS := base64.StdEncoding.EncodeToString(hash)
	v, _ := r.redisClient.Get(ctx, "hash:"+hashS).Result()

	return v == "1"
}

func (r redisDriver) Ensure(hash []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	hashS := base64.StdEncoding.EncodeToString(hash)
	r.redisClient.Set(ctx, "hash:"+hashS, "1", 0)
}

var _ hashverifier.HashCache = redisDriver{}

func newRedis(redisUrl string) (hashverifier.HashCache, error) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err = client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return redisDriver{redisClient: client}, nil
}
