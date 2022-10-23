package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/common-nighthawk/go-figure"
	"github.com/jakemakesstuff/spherical/db"
)

func displayVersion() {
	figure.NewFigure("Spherical", "rectangles", true).Print()
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
	if err = db.Migrate(); err != nil {
		panic(err)
	}

	// If this is just migrations, do not continue any further.
	if *migrationsOnly {
		return
	}

	// Generate or get the PGP key.
	_, privKeyArmored, err := db.GetPGPKey(func() (pubKey, privKey string) {
		fmt.Print("[pgp] Generating 4096-bit pgp key - this might take a while...")
		key, err := crypto.GenerateKey(
			"Spherical", "noreply.nodegen@spherical.gg", "rsa", 4096)
		if err != nil {
			panic(err)
		}
		pubKey, err = key.GetArmoredPublicKey()
		if err != nil {
			panic(err)
		}
		privKey, err = key.Armor()
		if err != nil {
			panic(err)
		}
		fmt.Println(" success!")
		return
	})
	if err != nil {
		panic(err)
	}

	// Load the PGP key in.
	key, err := crypto.NewKeyFromArmored(privKeyArmored)
	if err != nil {
		panic(err)
	}
	if !key.IsPrivate() {
		panic("key is not private")
	}
}
