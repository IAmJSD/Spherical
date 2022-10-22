package db

import (
	"context"
	"embed"
	"sort"
	"time"
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
	_, err = c.Exec(contextMaker(), "SELECT pg_advisory_lock('migrate') FROM migrations")
	if err != nil {
		return err
	}
	defer func() {
		_, _ = c.Exec(contextMaker(), "SELECT pg_advisory_unlock('migrate') FROM migrations")
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
		
	}
}
