package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jakemakesstuff/spherical/defaulttrusted"
	"github.com/jakemakesstuff/spherical/hashverifier"
	"github.com/valyala/fasthttp"
)

func getDriver(redisUrl, boltPathPtr *string) hashverifier.HashCache {
	if redisUrl != nil && *redisUrl != "" {
		driver, err := newRedis(*redisUrl)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to initialize redis:", err)
			os.Exit(1)
		}
		return driver
	}

	driver, err := newBolt(*boltPathPtr)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "failed to initialize boltdb:", err)
		os.Exit(1)
	}
	return driver
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "[item...]"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func normaliseSlice(a []string, hostname string) []string {
	newSlice := make([]string, 0, len(a))
	for _, v := range a {
		v = strings.ToLower(strings.TrimSpace(v))
		if v != "" && v != hostname {
			newSlice = append(newSlice, v)
		}
	}
	return newSlice
}

func main() {
	host := flag.String(
		"host", "0.0.0.0:8080", "the host that the hash validator will run on")
	redisUrl := flag.String(
		"redis-url", "", "defines the redis connection string if you wish to use this - overrides boltdb if set")
	boltPath := flag.String(
		"bolt-path", "./data", "defines the boltdb database path")
	consensus := flag.Uint(
		"consensus", 3, "defines the amount of trusted nodes required to override a invalid hash - set to 0 to disable")
	disableDefaultTrusted := flag.Bool(
		"disable-default-trusted", false, "disable the default trusted hash verification node list")
	hostname := flag.String(
		"hostnwme", "", "used to remove the hostname from the informants/trusted and set the skip header - should be set if node is default trusted")

	var informantsFlag arrayFlags
	flag.Var(
		&informantsFlag, "informants", "defines additional servers that should be informed of new hashes - will tell trusted anyway")
	var trustedFlag arrayFlags
	flag.Var(
		&trustedFlag, "trusted-nodes", "defines nodes that are trusted enough that <consensus> number of them can vouch that a node"+
			" signed something - useful if a node changes their pgp key")

	flag.Parse()

	if !*disableDefaultTrusted {
		// Inject the default trusted nodes into this.
		trustedFlag = append(trustedFlag, defaulttrusted.New()...)
	}

	informants := normaliseSlice(informantsFlag, *hostname)
	trusted := normaliseSlice(trustedFlag, *hostname)

	if *consensus > uint(len(trusted)) {
		_, _ = fmt.Fprintln(os.Stderr, "consensus count greater than trusted node count")
		os.Exit(1)
	}

	client := hashverifier.NewClient(
		getDriver(redisUrl, boltPath),
		informants,
		trusted,
		*hostname,
		*consensus)
	fmt.Println("Listening on host", *host)
	if err := fasthttp.ListenAndServe(*host, handler(client)); err != nil {
		panic(err)
	}
}
