package html

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const cacheTtl = time.Minute * 15

//go:embed template.html
var templateHtml []byte

var (
	globalTpl        *template.Template
	currentGlobalTpl []byte
	globalTplLock    = sync.RWMutex{}
)

// Handles trying to compile the global template. Returns it if successful.
func compileGlobalTemplate() (*template.Template, error) {
	globalTplLock.Lock()
	defer globalTplLock.Unlock()

	// Get the HTML.
	var html []byte
	var err error
	if tlruCache == nil {
		// The HTML should be got from the FS.
		html, err = os.ReadFile("html/template.html")
		if err != nil {
			return nil, err
		}
	} else {
		// The HTML should be got from the Go embed.
		html = templateHtml
	}

	globalTpl, err = template.New("base").Parse(string(html))
	if err != nil {
		return nil, err
	}
	currentGlobalTpl = html
	return globalTpl, nil
}

// Render is used to render the HTML to a byte slice or copy a cached version if it exists. Returns a byte slice.
func Render(title, description, publicUrl string, javascriptAssets []string, metadata any) (res []byte, err error) {
	// Get the additional HTML parts.
	additionalHead, additionalBody := getHtmlParts()

	// Encode the metadata. We will need it JSON encoded in both cache hit-and-miss cases.
	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	// Attempt to load from cache if this is in production.
	if tlruCache != nil {
		// Turn this into a big json blob.
		j, _ := json.Marshal([]any{
			additionalHead, additionalBody, title, description, publicUrl, javascriptAssets,
			json.RawMessage(metadataJson),
		})

		// Hash the sha256 blob.
		hasher := sha256.New()
		hasher.Write(j)
		hash := hasher.Sum(nil)

		// Check the cache.
		val := tlruCache.Get(string(hash), cacheTtl)
		if val != nil {
			// Copy the result.
			cpy := make([]byte, len(val))
			copy(cpy, val)
			return cpy, nil
		}

		// Write to the cache if there is a future success.
		defer func() {
			if err == nil && res != nil {
				// Write to the cache.
				cpy := make([]byte, len(res))
				copy(cpy, res)
				tlruCache.Set(string(hash), cpy, cacheTtl)
			}
		}()
	}

	// Get the template.
	var tpl *template.Template
	if tlruCache != nil {
		// Read lock the template and then try and get it if this is a cache hit.
		globalTplLock.RLock()
		if bytes.Equal(currentGlobalTpl, templateHtml) {
			// Set the template to the global one.
			tpl = globalTpl
		}
		globalTplLock.RUnlock()
	}

	// If the template is nil, handle compiling it.
	if tpl == nil {
		tpl, err = compileGlobalTemplate()
		if err != nil {
			// Compilation failed. Return here.
			return
		}
	}

	// Load/parse the manifest file.
	manifestFile, err := public.Open("bundles/manifest.json")
	if err != nil {
		// No manifest file for some reason.
		return
	}
	manifestJson, err := io.ReadAll(manifestFile)
	if err != nil {
		// Failed to read manifest file.
		return
	}
	var manifest map[string]string
	if err = json.Unmarshal(manifestJson, &manifest); err != nil {
		return
	}

	// Get the JS paths.
	jsPaths := make([]string, len(javascriptAssets))
	for i, v := range javascriptAssets {
		assetFile := manifest[v]
		if assetFile == "" {
			err = errors.New("asset file " + v + " does not exist")
			return
		}
		jsPaths[i] = "/bundles/" + assetFile
	}

	// Execute the template.
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, map[string]any{
		"Title":          title,
		"Description":    description,
		"PublicURL":      publicUrl,
		"JavaScripts":    jsPaths,
		"Metadata":       string(metadataJson),
		"AdditionalHead": template.HTML(additionalHead),
		"AdditionalBody": template.HTML(additionalBody),
	})
	res = buf.Bytes()

	// Return everything.
	return
}

// Handler is used make a handler to serve an HTTP response to an existing connection.
func Handler(
	title, description string, oobe bool, metadata any, statusCode int,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the appropriate JS bundle.
		bundles := []string{"main.js"}
		if oobe {
			bundles = []string{"oobe.js"}
		}

		// Get the public URL.
		// TODO: This is wrong because proxies.
		u := *r.URL
		u.RawPath = "/"
		u.Path = "/"
		u.RawQuery = ""
		publicUrl := u.String()

		// Do the render.
		content, err := Render(title, description, publicUrl, bundles, metadata)
		if err != nil {
			// Send the error to the console and set the content to an error.
			statusCode = 500
			_, _ = fmt.Fprintln(os.Stderr, "[http] Failed to make template for URL",
				r.URL, "-", err.Error())
			content = []byte(`<h1>Failed to compile template</h1>
<p>Please get the administrator to check the console to get the reason.</p>`)
		}

		// Write all the things.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(statusCode)
		_, _ = w.Write(content)
	}
}
