package db

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	currentPgConn    *pgxpool.Pool
	currentRedisConn *redis.Client
	currentConnLock  sync.RWMutex
)

func dbConn() *pgxpool.Pool {
	currentConnLock.RLock()
	defer currentConnLock.RUnlock()
	return currentPgConn
}

// Init is used to initialise the database connection.
func Init(pgConnectionUrl, redisConnectionUrl string) error {
	currentConnLock.Lock()
	defer currentConnLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	conn, err := pgxpool.Connect(ctx, pgConnectionUrl)
	if err != nil {
		return err
	}

	err = conn.Ping(ctx)
	if err != nil {
		return err
	}

	opts := &redis.Options{}
	if redisConnectionUrl != "" {
		opts, err = redis.ParseURL(redisConnectionUrl)
		if err != nil {
			return err
		}
	}
	redisClient := redis.NewClient(opts)
	err = redisClient.Ping(ctx).Err()
	if err != nil {
		return err
	}

	currentPgConn = conn
	currentRedisConn = redisClient
	return nil
}
