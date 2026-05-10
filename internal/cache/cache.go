package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user14923929/pak/internal/index"
)

const cacheDir = "/var/cache/pak"

// Save writes the package list for a given repo URL to disk.
func Save(repoURL string, pkgs []*index.Package) error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(cacheDir, safeFileName(repoURL)+".json")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(pkgs)
}

// Load reads all cached package lists and merges them.
func Load() ([]*index.Package, error) {
	entries, err := os.ReadDir(cacheDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var all []*index.Package
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		f, err := os.Open(filepath.Join(cacheDir, e.Name()))
		if err != nil {
			return nil, err
		}
		var pkgs []*index.Package
		if err := json.NewDecoder(f).Decode(&pkgs); err != nil {
			f.Close()
			return nil, fmt.Errorf("cache: corrupt file %s: %w", e.Name(), err)
		}
		f.Close()
		all = append(all, pkgs...)
	}
	return all, nil
}

// Search returns packages whose name contains the query string.
func Search(query string) ([]*index.Package, error) {
	all, err := Load()
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var out []*index.Package
	for _, p := range all {
		if strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Description), q) {
			out = append(out, p)
		}
	}
	return out, nil
}

// Find returns the package with exact name match, or nil.
func Find(name string) (*index.Package, error) {
	all, err := Load()
	if err != nil {
		return nil, err
	}
	for _, p := range all {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, nil
}

func safeFileName(url string) string {
	r := strings.NewReplacer("://", "_", "/", "_", ".", "_", ":", "_")
	return r.Replace(url)
}
