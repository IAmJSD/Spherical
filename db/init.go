package db

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	currentConn     *pgxpool.Pool
	currentConnLock sync.RWMutex
)

func conn() *pgxpool.Pool {
	currentConnLock.RLock()
	defer currentConnLock.RUnlock()
	return currentConn
}

// Init is used to initialise the database connection.
func Init(connectionUrl string) error {
	currentConnLock.Lock()
	defer currentConnLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	conn, err := pgxpool.Connect(ctx, connectionUrl)
	if err != nil {
		return err
	}

	err = conn.Ping(ctx)
	if err != nil {
		return err
	}

	currentConn = conn
	return nil
}
