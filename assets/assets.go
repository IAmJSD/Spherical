package assets

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/evanw/esbuild/pkg/api"
)

//go:embed **/* *
var assetsFs embed.FS

type metadata struct {
	StaticContent string            `json:"static_content"`
	JS            map[string]string `json:"js"`
	CSS           map[string]string `json:"css"`
}

// Result is the result of the compilation.
type Result struct {
	// JS maps the bundle name to the HTTP path for it.
	JS map[string]string `json:"js"`

	// JSSourceMap maps the bundle name to the HTTP path for it.
	JSSourceMap map[string]string `json:"js_source_map"`

	// CSS maps the bundle name to the HTTP path for it.
	CSS map[string]string `json:"css"`

	// CSSSourceMap maps the bundle name to the HTTP path for it.
	CSSSourceMap map[string]string `json:"CSS_source_map"`

	// FS is the returned filesystem with all of this injected into it.
	FS fs.FS `json:"fs"`
}

func copyFile(fs fs.ReadDirFS, v os.DirEntry, prefix, fsPath string) error {
	// Pogchamp! Write the file into the folder.
	newFsPath := path.Join(prefix, v.Name())
	f, err := fs.Open(newFsPath)
	if err != nil {
		// Unable to turn the path into a file.
		return err
	}
	defer f.Close()

	newDiskPath := filepath.Join(fsPath, v.Name())
	osFile, err := os.Create(newDiskPath)
	if err != nil {
		// Unable to make OS file.
		return err
	}
	defer osFile.Close()

	b := make([]byte, 1024)
	for {
		n, err := f.Read(b)
		if err != nil {
			if errors.Is(err, io.EOF) {
				// EOF. We are done.
				return nil
			}
			return err
		}
		b := b[:n]
		writeN, err := osFile.Write(b)
		if err != nil {
			return err
		}
		if writeN != n {
			return io.EOF
		}
	}
}

// A unideal hack to ensure that the filesystem contents are mapped properly.
func copyFromFsToDisk(fs fs.ReadDirFS, prefix, fsPath string) error {
	// Get the directory.
	dir, err := fs.ReadDir(prefix)
	if err != nil {
		return err
	}

	for _, v := range dir {
		// Check if this is a directory. Directories should recursively call themselves.
		if v.IsDir() {
			newDir := filepath.Join(fsPath, v.Name())
			if err := os.Mkdir(newDir, 0o777); err != nil {
				// Failed to make the sub-directory.
				return err
			}
			if err := copyFromFsToDisk(fs, path.Join(prefix, v.Name()), newDir); err != nil {
				// The child failed for some reason. Return their error.
				return err
			}
		} else {
			// Do the file copy.
			if err := copyFile(fs, v, prefix, fsPath); err != nil {
				return err
			}
		}
	}

	return nil
}

type fallThroughFs struct {
	css map[string][]byte
	js  map[string][]byte
	fs  fs.FS
}

var (
	cssRegexp = regexp.MustCompile(`^\.?/?css/`)
	jsRegexp  = regexp.MustCompile(`^\.?/?js/`)
)

type fakeFile struct {
	r *bytes.Reader

	b        []byte
	filename string
}

type staticStat struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
	sys     any
}

func (s staticStat) Name() string { return s.name }

func (s staticStat) Size() int64 { return s.size }

func (s staticStat) Mode() fs.FileMode { return s.mode }

func (s staticStat) ModTime() time.Time { return s.modTime }

func (s staticStat) IsDir() bool { return s.isDir }

func (s staticStat) Sys() any { return nil }

func (f fakeFile) Stat() (fs.FileInfo, error) {
	return staticStat{
		name:    f.filename,
		size:    int64(len(f.b)),
		mode:    0o777,
		modTime: time.Now(),
		isDir:   false,
		sys:     nil,
	}, nil
}

func (f fakeFile) Read(b []byte) (int, error) {
	return f.r.Read(b)
}

func (f fakeFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.r.ReadAt(p, off)
}

func (f fakeFile) Seek(offset int64, whence int) (int64, error) {
	return f.r.Seek(offset, whence)
}

var (
	_ io.Reader   = fakeFile{}
	_ io.ReaderAt = fakeFile{}
	_ io.Seeker   = fakeFile{}
)

func (f fakeFile) Close() error { return nil }

func newFakeFile(b []byte, filename string) fs.File {
	return fakeFile{
		r:        bytes.NewReader(b),
		b:        b,
		filename: filename,
	}
}

func (f fallThroughFs) Open(name string) (fs.File, error) {
	if cssRegexp.MatchString(name) {
		_, filename := path.Split(name)
		if b, ok := f.css[filename]; ok {
			// Return this.
			return newFakeFile(b, filename), nil
		}
	}

	if jsRegexp.MatchString(name) {
		_, filename := path.Split(name)
		if b, ok := f.js[filename]; ok {
			// Return this.
			return newFakeFile(b, filename), nil
		}
	}

	return f.fs.Open(name)
}

var _ fs.FS = fallThroughFs{}

// CompilationResult is used to define the result of the compilation.
var CompilationResult Result

// Init is used to initialise the asset compilation.
func Init() error {
	// Load the metadata file.
	metadataBytes, err := assetsFs.ReadFile("metadata.json")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "[assets] Failed to read metadata.json")
		return err
	}
	var m metadata
	err = json.Unmarshal(metadataBytes, &m)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "[assets] Failed to parse metadata.json")
		return err
	}
	fmt.Println("[assets] Found and parsed metadata.json!")

	// Write the JS folder to disk.
	jsFolder, err := os.MkdirTemp("", "spherical_js_*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(jsFolder) }()
	if err := copyFromFsToDisk(assetsFs, "content/js", jsFolder); err != nil {
		return err
	}

	// Go ahead and build the bundles.
	returnedJsFiles := map[string][]byte{}
	returnedCSSFiles := map[string][]byte{}
	fsSub, _ := fs.Sub(assetsFs, "content/static")
	results := Result{
		JS:           map[string]string{},
		JSSourceMap:  map[string]string{},
		CSS:          map[string]string{},
		CSSSourceMap: map[string]string{},
		FS: fallThroughFs{
			css: returnedCSSFiles,
			js:  returnedJsFiles,
			fs:  fsSub,
		},
	}

	// Process the JS bundles.
	for bundleName, entrypoint := range m.JS {
		// Compiles the JS bundle.
		fmt.Print("[assets] Compiling JS asset " + bundleName + "...")
		result := api.Build(api.BuildOptions{
			EntryPoints:       []string{entrypoint},
			AbsWorkingDir:     jsFolder,
			Outfile:           filepath.Join(jsFolder, "out.js"),
			Sourcemap:         api.SourceMapExternal,
			MinifyIdentifiers: true,
			MinifySyntax:      true,
			MinifyWhitespace:  true,
			Target:            api.ES2020,
			Bundle:            true,
		})
		if len(result.Errors) != 0 {
			errorStrings := []string{}
			for _, v := range result.Errors {
				errorStrings = append(errorStrings, v.Text)
			}
			return errors.New(strings.Join(errorStrings, ", "))
		}
		fmt.Println(" success!")

		// Get the files and load them into memory.
		for _, v := range result.OutputFiles {
			// Get the filename.
			_, filename := filepath.Split(v.Path)

			// Delete the file right off the bat.
			_ = os.Remove(v.Path)

			// Make a hash of the file.
			sha := sha256.New()
			sha.Write(v.Contents)
			hash := base64.URLEncoding.EncodeToString(sha.Sum(nil))

			// Put this into the appropriate place.
			fsFilename := bundleName + "." + hash + ".js"
			if strings.HasSuffix(filename, ".map") {
				fsFilename += ".map"
				returnedJsFiles[fsFilename] = v.Contents
				results.JSSourceMap[bundleName] = "/assets/js/" + fsFilename
			} else {
				returnedJsFiles[fsFilename] = v.Contents
				results.JS[bundleName] = "/assets/js/" + fsFilename
			}
		}
	}

	// Write the CSS folder to disk.
	CSSFolder, err := os.MkdirTemp("", "spherical_CSS_*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(CSSFolder) }()
	if err := copyFromFsToDisk(assetsFs, "content/css", CSSFolder); err != nil {
		return err
	}

	// Process the CSS bundles.
	for bundleName, entrypoint := range m.CSS {
		// Compiles the CSS bundle.
		fmt.Print("[assets] Compiling CSS asset " + bundleName + "...")
		result := api.Build(api.BuildOptions{
			EntryPoints:       []string{entrypoint},
			AbsWorkingDir:     CSSFolder,
			Outfile:           filepath.Join(jsFolder, "out.css"),
			Sourcemap:         api.SourceMapExternal,
			MinifyIdentifiers: true,
			MinifySyntax:      true,
			MinifyWhitespace:  true,
			Bundle:            true,
		})
		if len(result.Errors) != 0 {
			errorStrings := []string{}
			for _, v := range result.Errors {
				errorStrings = append(errorStrings, v.Text)
			}
			return errors.New(strings.Join(errorStrings, ", "))
		}
		fmt.Println(" success!")

		// Get the files and load them into memory.
		for _, v := range result.OutputFiles {
			// Get the filename.
			_, filename := filepath.Split(v.Path)

			// Delete the file right off the bat.
			_ = os.Remove(v.Path)

			// Make a hash of the file.
			sha := sha256.New()
			sha.Write(v.Contents)
			hash := base64.URLEncoding.EncodeToString(sha.Sum(nil))

			// Put this into the appropriate place.
			fsFilename := bundleName + "." + hash + ".css"
			if strings.HasSuffix(filename, ".map") {
				fsFilename += ".map"
				returnedCSSFiles[fsFilename] = v.Contents
				results.CSSSourceMap[bundleName] = "/assets/css/" + fsFilename
			} else {
				returnedCSSFiles[fsFilename] = v.Contents
				results.CSS[bundleName] = "/assets/css/" + fsFilename
			}
		}
	}

	fmt.Println("[assets] Successfully built the following JS:", results.JS)
	fmt.Println("[assets] Successfully built the following CSS:", results.CSS)
	CompilationResult = results
	return nil
}
