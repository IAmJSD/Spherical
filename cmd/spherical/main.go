package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jakemakesstuff/spherical/db"
)

func displayVersion() {

}

func main() {
	// Get the flags.
	postgresUrl := flag.String("postgres-url", os.Getenv("POSTGRES_URL"),
		"the postgres connection url - defaults to the POSTGRES_URL env variable")
	migrationsOnly := flag.Bool("migrations-only", false, "only run migrations and then return")
	flag.Parse()

	// Print the version.
	displayVersion()

	// Connect to the database.
	fmt.Print("[db] Connecting to postgres...")
	err := db.Init(*postgresUrl)
	if err != nil {
		panic(err)
	}
	fmt.Println(" success!")

	// Run the database migrations.
	err = db.Migrate()
	if err != nil {
		panic(err)
	}

	// If this is just migrations, do not continue any further.
	if *migrationsOnly {
		return
	}
}
