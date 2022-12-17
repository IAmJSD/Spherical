package hashverifier

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/jakemakesstuff/spherical/utils/random"
)

var rand = random.New()

// Client is used to define a client for hash verification. This is important for checking a message actually cams from
// a node as well as keeping a record of it being sent on many servers.
type Client struct {
	// Defines the cache implementation.
	cache HashCache

	// Defines hash verification servers this client should inform of new hashes sent its way. Note that it also sends
	// to trusted servers if they do not overlap with the informants.
	informants     []string
	informantsLock sync.RWMutex

	// Defines how much consensus on how many trusted nodes have to be reached before an unverifiable hash becomes
	// verified in this server. If this is set to zero, it will turn off informant caching. By default, this should
	// be set to 3.
	consensus uint

	// Defines our current hostname.
	hostname string

	// Defines servers that are trusted. A trusted server is used alongside the consensus logic to determine if a hash
	// was valid at some point (this might be the case if for example the server was deleted, the key was changed, the
	// server was pretending it never sent a message).
	trusted     []string
	trustedLock sync.RWMutex
}

// NewClient is used to make a new version of the client above.
func NewClient(cache HashCache, informants, trusted []string, hostname string, consensus uint) *Client {
	if cache == nil {
		cache = nopCache{}
	}
	for i, v := range informants {
		informants[i] = strings.ToLower(v)
	}
	for i, v := range trusted {
		trusted[i] = strings.ToLower(v)
	}
	return &Client{
		cache:      cache,
		informants: informants,
		consensus:  consensus,
		hostname:   strings.TrimSpace(strings.ToLower(hostname)),
		trusted:    trusted,
	}
}

// AddInformants is used to add informants to the slice.
func (c *Client) AddInformants(informants ...string) {
	c.informantsLock.Lock()
	defer c.informantsLock.Unlock()
	for _, informant := range informants {
		informant = strings.ToLower(informant)
		found := false
		for _, v := range c.informants {
			if v == informant {
				found = true
				break
			}
		}
		if found {
			// Informant already within app. Continue.
			continue
		}
		c.informants = append(c.informants, informant)
	}
}

// AddTrustedNodes is used to add trusted nodes to the slice.
func (c *Client) AddTrustedNodes(nodes ...string) {
	c.trustedLock.Lock()
	defer c.trustedLock.Unlock()
	for _, node := range nodes {
		node = strings.ToLower(node)
		found := false
		for _, v := range c.trusted {
			if v == node {
				found = true
				break
			}
		}
		if found {
			// Node already within app. Continue.
			continue
		}
		c.trusted = append(c.trusted, node)
	}
}

func lookupPgp(ctx context.Context, hostname string) []byte {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	u, _ := url.Parse("https://example.com/spherical.pub")
	u.Host = hostname
	urlStr := u.String()

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		// Something went wrong making the request. Probably a bad url.
		return nil
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// Failed to make HTTP request.
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		// Bad response code.
		return nil
	}

	b, err := io.ReadAll(io.LimitReader(res.Body, 1000000))
	if err != nil {
		// Unable to read body.
		return nil
	}
	return b
}

func (c *Client) getInformantsAndTrusted() []string {
	// Make the set.
	set := map[string]bool{}
	c.informantsLock.RLock()
	for _, v := range c.informants {
		set[v] = true
	}
	c.informantsLock.RUnlock()
	c.trustedLock.RLock()
	for _, v := range c.trusted {
		set[v] = true
	}
	c.trustedLock.RUnlock()

	// Make the slice.
	s := make([]string, len(set))
	i := 0
	for v := range set {
		s[i] = v
		i++
	}
	return s
}

var trueB = []byte("true")

func (c *Client) mergeWithTrusted(skip []string) []string {
	trusted := c.trusted
	l := len(trusted) + len(skip)
	if c.hostname != "" {
		l++
	}
	a := make([]string, 1)
	copy(a, trusted)
	copy(a[len(trusted):], skip)
	if c.hostname != "" {
		a[l-1] = c.hostname
	}
	return a
}

// ProcessHashBlob is used to process a hash blob that was sent to you. Returns the boolean that should be returned
// to the user.
func (c *Client) ProcessHashBlob(ctx context.Context, blob string, skip []string) bool {
	// Make sure it doesn't end with a new line.
	blob = strings.TrimSpace(blob)

	// Handle skip hostnames.
	if skip == nil || c.hostname == "" {
		for _, v := range skip {
			if v == c.hostname {
				// Skip this host.
				return false
			}
		}
	}

	// Hash all of this together.
	sha := sha256.New()
	_, _ = sha.Write([]byte(blob))
	totalHash := sha.Sum(nil)

	// Before we do anything further, check if we have already dealt with this in the cache.
	if c.cache.Exists(totalHash) {
		// If we have, just return true right away. We do not need to worry about informing, we would've done that on
		// send.
		return true
	}

	// If we don't, we need to verify it. Start by splitting the hostname, hash, and PGP signature.
	blobSplit := strings.SplitN(blob, "\n", 3)
	hostname := blobSplit[0]
	hash := blobSplit[1]

	// Get the PGP signature.
	pgpSignature, err := crypto.NewPGPSignatureFromArmored(blobSplit[2])
	if err != nil {
		// This cannot be a valid message.
		return false
	}

	// Check if we can get the PGP public key.
	pgpPublicKey := c.cache.LookupPGP(hostname)
	webLookupRan := false
internetLookup:
	if pgpPublicKey == nil {
		// Try and get the PGP key from the internet.
		webLookupRan = true
		if pgpPublicKey = lookupPgp(ctx, hostname); pgpPublicKey != nil {
			// We found a PGP key. Write it into the cache.
			c.cache.WritePGP(hostname, pgpPublicKey)
		}
	}

	// Read the PGP key if it exists.
	if pgpPublicKey != nil {
		key, err := crypto.NewKeyFromArmored(string(pgpPublicKey))
		if err == nil {
			// Get the keyring.
			signingKeyring, err := crypto.NewKeyRing(key)
			if err == nil {
				// Check if this is valid.
				if err = signingKeyring.VerifyDetached(
					crypto.NewPlainMessage([]byte(hash)),
					pgpSignature,
					crypto.GetUnixTime(),
				); err == nil {
					// This hash was successfully verified to have come from this server by us. Tell our cache driver,
					// any informants/trusted servers, and return true.
					c.cache.Ensure(totalHash)
					InformVerifiers(
						hostname,
						hash,
						blobSplit[2],
						c.getInformantsAndTrusted())
					return true
				}
			}
		}

		// Check if we have done a web lookup. If not, nil the key and jump back to that bit.
		if !webLookupRan {
			pgpPublicKey = nil
			goto internetLookup
		}
	}

	c.trustedLock.RLock()
	if c.consensus != 0 && len(c.trusted) >= int(c.consensus) {
		// Make a copy of the slice.
		trusted := make([]string, len(c.trusted))
		copy(trusted, c.trusted)
		c.trustedLock.RUnlock()

		// Shuffle the trusted slice.
		rand.Shuffle(len(trusted), func(i, j int) {
			trusted[i], trusted[j] = trusted[j], trusted[i]
		})

		// Get the number of concurrent requests. Should be 3 or consensus count, whichever is higher.
		concurrentCount := 3
		if int(c.consensus) > concurrentCount {
			concurrentCount = int(c.consensus)
		}

		// Start making requests to trusted nodes.
		var returnedTrue uint64
		for atomic.LoadUint64(&returnedTrue) >= uint64(c.consensus) || len(trusted) != 0 {
			// Take a number of items from the slice.
			take := concurrentCount
			trustedLen := len(trusted)
			if take > trustedLen {
				take = trustedLen
			}
			servers := trusted[:take]
			trusted = trusted[take:]

			// Get the results of all nodes.
			wg := sync.WaitGroup{}
			wg.Add(len(servers))
			for _, node := range servers {
				// If it is in skip, skip it.
				if skip != nil {
					skipNode := false
					for _, v := range skip {
						if v == node {
							skipNode = true
							break
						}
					}
					if skipNode {
						wg.Done()
						continue
					}
				}

				// Start a goroutine to deal with it.
				node := node
				go func() {
					defer wg.Done()

					u, err := url.Parse("https://example.com/verify")
					if err != nil {
						return
					}
					u.Host = node

					req, err := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(blob))
					if err != nil {
						return
					}

					req.Header.Set("X-Skip", strings.Join(c.mergeWithTrusted(skip), ","))

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != 200 {
						return
					}

					body, err := io.ReadAll(io.LimitReader(resp.Body, 20))
					if err == nil && bytes.Equal(body, trueB) {
						atomic.AddUint64(&returnedTrue, 1)
					}
				}()
			}
			wg.Wait()
		}

		// Return if there is a consensus.
		return uint(returnedTrue) >= c.consensus
	} else {
		// Just immediately unlock.
		c.trustedLock.RUnlock()
	}

	// No consensus reached.
	return false
}
