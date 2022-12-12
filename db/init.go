package db

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	currentPgConn    *pgxpool.Pool
	currentRedisConn *redis.Client
	backgroundCancel func()
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

	if backgroundCancel != nil {
		// Cancel the current background context and then nil it to make sure it is never called again.
		backgroundCancel()
		backgroundCancel = nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	conn, err := pgxpool.New(ctx, pgConnectionUrl)
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

	var backgroundCtx context.Context
	backgroundCtx, backgroundCancel = context.WithCancel(context.Background())
	go watchLoop(backgroundCtx, currentRedisConn)

	return nil
}
