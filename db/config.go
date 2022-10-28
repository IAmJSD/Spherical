package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgtype"
)

type configChange struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

// InternallyConsumeConfig is used to getting the config and handling events relating to it. This will block until the config is
// initially fetched, but will spin off as a side effect logic to auto-update it. The handler SHOULD be okay with concurrency.
// Note you probably do not want this function, but instead functions exposed in the config package.
func InternallyConsumeConfig(handler func(key string, value json.RawMessage)) error {
	// Add event to watch for new changes.
	AddWatchEvent("config", func(c configChange) { handler(c.Key, c.Value) })

	// Make the database query for the config.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	rows, err := dbConn().Query(ctx, "SELECT key, value FROM config")
	if err != nil {
		return err
	}
	defer rows.Close()

	// Handle each row.
	for rows.Next() {
		// Get the key and value.
		var key string
		var value pgtype.JSONB
		if err := rows.Scan(&key, &value); err != nil {
			return err
		}

		// Get the raw JSON.
		var b json.RawMessage
		if err := value.AssignTo(&b); err != nil {
			return err
		}

		// Call the function handler.
		handler(key, b)
	}

	// Return no errors.
	return nil
}

// UpdateConfig is used to write a key to the database and also dispatch events for it. Blocks until events
// are done.
func UpdateConfig(ctx context.Context, key string, value any) error {
	valueMarshal, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return EditAndPublish(ctx, "config", func(ctx context.Context) error {
		query := "INSERT INTO config (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2"
		_, err := dbConn().Exec(ctx, query, key, pgtype.JSONB{Bytes: valueMarshal, Status: pgtype.Present})
		return err
	}, configChange{Key: key, Value: valueMarshal})
}
