package db

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate runs database migrations.
func Migrate() error {
	// Get the migrations.
	entries, readErr := migrations.ReadDir("migrations")
	if readErr != nil {
		return readErr
	}
	filenames := make([]string, len(entries))
	for i, v := range entries {
		filenames[i] = v.Name()
	}
	sort.Strings(filenames)

	// Handle the function to make all contexts.
	cancellers := []func(){}
	defer func() {
		for _, v := range cancellers {
			v()
		}
	}()
	contextMaker := func() context.Context {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		cancellers = append(cancellers, cancel)
		return ctx
	}

	// Ensure the migrations table exists.
	c := dbConn()
	_, tableErr := c.Exec(contextMaker(), "CREATE TABLE IF NOT EXISTS migrations (filename TEXT PRIMARY KEY)")
	if tableErr != nil {
		return tableErr
	}

	// Get the lock.
	_, err := UseGlobalLock(contextMaker(), "migrations", func() (struct{}, error) {
		// Get all the migrations from the table.
		rows, migrationsGetErr := c.Query(contextMaker(), "SELECT filename FROM migrations")
		if migrationsGetErr != nil {
			return struct{}{}, migrationsGetErr
		}
		defer rows.Close()
		migrationsRan := make([]string, 0)
		for rows.Next() {
			var filename string
			if migrationsGetErr = rows.Scan(&filename); migrationsGetErr != nil {
				return struct{}{}, migrationsGetErr
			}
			migrationsRan = append(migrationsRan, filename)
		}
		rows.Close()

		// Run all pending migrations.
		for _, filename := range filenames {
			// Check if it has already been ran.
			ran := false
			for _, dbFilename := range migrationsRan {
				if dbFilename == filename {
					fmt.Println("[db] Migration", filename, "already ran - skipping!")
					ran = true
					break
				}
			}

			// Run the migration if not.
			if !ran {
				// Get the migration SQL.
				migrationSql, err := migrations.ReadFile("migrations/" + filename)
				if err != nil {
					return struct{}{}, err
				}

				// Check if it starts with "-- nosplit".
				parts := [][]byte{}
				if bytes.HasPrefix(migrationSql, []byte("-- nosplit")) {
					parts = [][]byte{migrationSql}
				} else {
					parts = bytes.Split(migrationSql, []byte(";"))
				}
				batch := &pgx.Batch{}
				for _, v := range parts {
					batch.Queue(string(v))
				}
				batch.Queue("INSERT INTO migrations (filename) VALUES ($1)", filename)
				fmt.Print("[db] Running migration ", filename, "...")
				results := c.SendBatch(contextMaker(), batch)
				for i := 0; i < len(parts)+1; i++ {
					_, err = results.Exec()
					if err != nil && err.Error() != "no result" {
						_ = results.Close()
						return struct{}{}, err
					}
				}
				_ = results.Close()
				fmt.Println(" success!")
			}
		}

		// Return no errors.
		fmt.Println("[db] All migrations ran!")
		return struct{}{}, nil
	})
	return err
}
