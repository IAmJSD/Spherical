package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// InternallyDeleteToken is used to internally delete a token. This is used by the token delete job.
// Returns the deletion job ID and any errors.
func InternallyDeleteToken(ctx context.Context, token string) (string, error) {
	query := "DELETE FROM sessions WHERE token = $1 RETURNING deletion_job_id"
	var deletionJobID string
	err := dbConn().QueryRow(ctx, query, token).Scan(&deletionJobID)
	return deletionJobID, err
}

// InternallyGetTokenJobID is used to get the job ID of a token. This is used by the token extend job.
func InternallyGetTokenJobID(ctx context.Context, token string) (string, error) {
	query := "SELECT deletion_job_id FROM sessions WHERE token = $1"
	var deletionJobID string
	err := dbConn().QueryRow(ctx, query, token).Scan(&deletionJobID)
	return deletionJobID, err
}

// BuildCrossNodeToken is used to build a 2 min TTL token which can be used to connect to the gateway.
func BuildCrossNodeToken(ctx context.Context, data []byte) (string, error) {
	// Get the Redis connection.
	currentConnLock.RLock()
	redisConn := currentRedisConn
	currentConnLock.RUnlock()

	// Get the UUID.
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	uidS := uid.String()

	// Write the data to Redis.
	_, err = redisConn.Set(ctx, uidS, data, 2*time.Minute).Result()
	return uidS, err
}

// GetCrossNodeToken is used to get the data from a cross node token and destroy it at the same time.
func GetCrossNodeToken(ctx context.Context, token string) ([]byte, error) {
	// Get the Redis connection.
	currentConnLock.RLock()
	redisConn := currentRedisConn
	currentConnLock.RUnlock()

	// Acquire a lock to ensure we do not race on getting and deleting tokens.
	c := dbConn()
	_, err := c.Exec(ctx, "SELECT pg_advisory_lock(1221)")
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = c.Exec(ctx, "SELECT pg_advisory_unlock(1221)")
	}()

	// Get the data from Redis.
	data, err := redisConn.Get(ctx, token).Bytes()
	if err != nil {
		return nil, err
	}

	// Delete the data from Redis.
	_, err = redisConn.Del(ctx, token).Result()
	return data, err
}
