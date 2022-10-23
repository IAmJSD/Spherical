package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

// GetPGPKey is used to get the PGP key or create it if it doesn't exist.
func GetPGPKey(maker func() (pubKey, privKey string)) (pubKey, privKey string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer func() {
		// In a separate function to stop any references breaking below.
		cancel()
	}()

	// Make sure the pgp table exists.
	c := dbConn()
	_, err = c.Exec(ctx, "CREATE TABLE IF NOT EXISTS pgp_key (pub_key TEXT NOT NULL, priv_key TEXT NOT NULL)")
	if err != nil {
		return
	}

	// Get the advisory lock.
	_, err = c.Exec(ctx, "SELECT pg_advisory_lock(1010)")
	if err != nil {
		return
	}
	defer func() {
		_, _ = c.Exec(ctx, "SELECT pg_advisory_unlock(1010)")
	}()

	// Check if there is a pgp key.
	fmt.Print("[db] Looking for pgp key...")
	row := c.QueryRow(ctx, "SELECT pub_key, priv_key FROM pgp_key")
	if err = row.Scan(&pubKey, &privKey); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			// Some other error not relating to it not existing.
			return
		}

		// Set the error to nil and log.
		err = nil
		fmt.Println(" not found!")

		// Call the maker.
		pubKey, privKey = maker()

		// Make a new context. It is VERY likely 2 seconds have passed by now.
		cancel()
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*2)

		// Write the PGP key to the database.
		fmt.Print("[db] Writing new pgp key...")
		_, err = c.Exec(ctx, "INSERT INTO pgp_key (pub_key, priv_key) VALUES ($1, $2)", pubKey, privKey)
		if err != nil {
			return
		}
		fmt.Println(" success!")

		// Return here.
		return
	}
	fmt.Println(" found!")
	return
}
