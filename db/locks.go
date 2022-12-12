package db

import (
	"context"

	farmhash "github.com/leemcloughlin/gofarmhash"
)

// UseGlobalLock uses a global lock until fn is done. The value is passed through from the function and any errors
// are returned.
func UseGlobalLock[T any](ctx context.Context, lockName string, fn func() (T, error)) (val T, err error) {
	lockNumber := int64(farmhash.Hash64([]byte(lockName)))
	c := dbConn()
	_, err = c.Exec(ctx, "SELECT pg_advisory_lock($1)", lockNumber)
	if err != nil {
		return
	}
	val, err = fn()
	_, lockErr := c.Exec(ctx, "SELECT pg_advisory_unlock($1)", lockNumber)
	if err == nil {
		err = lockErr
	}
	return
}
