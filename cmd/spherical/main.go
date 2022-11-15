package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/common-nighthawk/go-figure"
	packagejson "github.com/jakemakesstuff/spherical"
	"github.com/jakemakesstuff/spherical/config"
	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/html"
	"github.com/jakemakesstuff/spherical/httproutes"
)

func displayVersion() {
	figure.NewFigure("Spherical", "rectangles", true).Print()
	p := packagejson.Package()
	fmt.Println("Version", p.Version, "\n---------------------------")
}

func oneOf(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func boolVal(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "yes" || s == "true" || s == "1"
}

func main() {
	// Get the flags.
	postgresUrl := flag.String("postgres-url", os.Getenv("POSTGRES_URL"),
		"the postgres connection url - defaults to the POSTGRES_URL env variable")
	redisUrl := flag.String("redis-url", os.Getenv("REDIS_URL"),
		"the redis connection url - defaults to the REDIS_URL env variable")
	listener := flag.String("listen-addr", oneOf(os.Getenv("LISTEN_ADDR"), "127.0.0.1:5000"),
		"defines the string to listen on, Defaults to the LISTEN_ADDR env variable, or if that is not present, "+
			"127.0.0.1:5000.")
	migrationsOnly := flag.Bool("migrations-only", false, "only run migrations and then return")
	flag.Parse()

	// Print the version.
	displayVersion()

	// Setup the HTML templater.
	isDev := boolVal(os.Getenv("DEV"))
	html.Setup(isDev)

	// Connect to the database.
	fmt.Print("[db] Connecting to postgres and redis...")
	err := db.Init(*postgresUrl, *redisUrl)
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

	// Start watching the config.
	if err := config.Watch(); err != nil {
		panic(err)
	}

	// Get the setup key if this is not already configured.
	if !config.Config().Setup {
		// Get the setup key and repeat it in the console.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		key, err := db.SetupKey(ctx)
		if err != nil {
			panic(err)
		}
		logMessage := "[setup] Your setup key is " + key + " - repeating 3 times!"
		for i := 0; i < 3; i++ {
			fmt.Println(logMessage)
		}
	}

	// Start the listener.
	fmt.Println("[http] Starting listener on", *listener)
	if err := http.ListenAndServe(*listener, httproutes.SelectRouter(isDev)); err != nil {
		panic(err)
	}
}
