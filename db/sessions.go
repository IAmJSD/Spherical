package db

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
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

// InternallyDeleteAllUserIDTokens is used to delete all tokens for a user ID. Returns the deletion job ID's.
func InternallyDeleteAllUserIDTokens(ctx context.Context, userId uint64) ([]string, error) {
	// Handle deleting session tokens.
	query := "DELETE FROM sessions WHERE user_id = $1 RETURNING deletion_job_id"
	rows, err := dbConn().Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	s := []string{}
	for rows.Next() {
		var deletionJobID string
		err = rows.Scan(&deletionJobID)
		if err != nil {
			return nil, err
		}
		s = append(s, deletionJobID)
	}

	// Handle deleting half tokens.
	currentConnLock.RLock()
	redisConn := currentRedisConn
	currentConnLock.RUnlock()
	_, err = redisConn.SRem(ctx, "half_tokens:"+strconv.FormatUint(userId, 10)).Result()
	if err != nil {
		return nil, err
	}

	// Return the jobs.
	return s, nil
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
	return UseGlobalLock(ctx, "cross_node_token:"+token, func() ([]byte, error) {
		// Get the data from Redis.
		data, err := redisConn.Get(ctx, token).Bytes()
		if err != nil {
			return nil, err
		}

		// Delete the data from Redis.
		_, err = redisConn.Del(ctx, token).Result()
		return data, err
	})
}

// BuildHalfToken is used to build a half token which can be used to link to a user ID and prove the first stage of
// authentication was successful.
func BuildHalfToken(ctx context.Context, userId uint64) (string, error) {
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

	// Write the half token -> user ID to redis.
	_, err = redisConn.Set(ctx, "half_token:"+uidS, strconv.FormatUint(userId, 10), time.Minute*30).Result()
	if err != nil {
		return "", err
	}

	// Add to the Redis user index.
	_, err = redisConn.SAdd(ctx, "half_tokens:"+strconv.FormatUint(userId, 10), uidS).Result()
	return uidS, err
}

// GetHalfToken is used to get the user ID from a half token and destroy it at the same time. If the user doesn't exist,
// 0 is returned.
func GetHalfToken(ctx context.Context, token string) (uint64, error) {
	// Get the Redis connection.
	currentConnLock.RLock()
	redisConn := currentRedisConn
	currentConnLock.RUnlock()

	// Acquire a lock to ensure we do not race on getting and deleting tokens.
	return UseGlobalLock(ctx, "half_token:"+token, func() (uint64, error) {
		// Get the user ID from Redis.
		userIdS, err := redisConn.Get(ctx, "half_token:"+token).Result()
		if err != nil {
			// Check if the error is that the key does not exist. If it is, just set the error to nil.
			if errors.Is(err, redis.Nil) {
				err = nil
			}

			// Return the result.
			return 0, err
		}
		userId, err := strconv.ParseUint(userIdS, 10, 64)
		if err != nil {
			return 0, err
		}

		// Delete the half token from Redis.
		_, err = redisConn.Del(ctx, "half_token:"+token).Result()
		if err != nil {
			return 0, err
		}

		// Remove the half token from the user index.
		res, err := redisConn.SRem(ctx, "half_tokens:"+userIdS, token).Result()
		if err != nil {
			return 0, err
		}
		if res == 0 {
			// This means that there has been a situation where the user has logged out or forcefully had their session
			// cleared, but the half token was not deleted. This is fine because we can keep track of it with this
			// index, but we effectively need to play dead now.
			return 0, nil
		}
		return userId, nil
	})
}

// CreateSession is used to create a new session for a user ID.
func CreateSession(ctx context.Context, token string, userId uint64, jobId string) error {
	query := "INSERT INTO sessions (token, user_id, deletion_job_id) VALUES ($1, $2, $3)"
	_, err := dbConn().Exec(ctx, query, token, userId, jobId)
	return err
}
