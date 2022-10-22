package db

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v4"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate runs database migrations.
func Migrate() error {
	// Get the migrations.
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		panic(err)
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
	c := conn()
	_, err = c.Exec(contextMaker(), "CREATE TABLE IF NOT EXISTS migrations (filename TEXT PRIMARY KEY)")
	if err != nil {
		return err
	}

	// Acquire a lock on the table.
	_, err = c.Exec(contextMaker(), "SELECT pg_advisory_lock(69420)")
	if err != nil {
		return err
	}
	defer func() {
		_, _ = c.Exec(contextMaker(), "SELECT pg_advisory_unlock(69420)")
	}()

	// Get all the migrations from the table.
	rows, err := c.Query(contextMaker(), "SELECT filename FROM migrations")
	if err != nil {
		return err
	}
	defer rows.Close()
	migrationsRan := make([]string, 0)
	for rows.Next() {
		var filename string
		if err = rows.Scan(&filename); err != nil {
			return err
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
				return err
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
					panic(err)
				}
			}
			_ = results.Close()
			fmt.Println(" success!")
		}
	}

	// Return no errors.
	fmt.Println("[db] All migrations ran!")
	return nil
}
