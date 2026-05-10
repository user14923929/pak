package fetch

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Get downloads url into a temp file and returns its path.
// If wantSHA256 is non-empty, the checksum is verified.
func Get(url, wantSHA256 string) (string, error) {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %s: HTTP %d", url, resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "pak-*.tmp")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, h), resp.Body); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("fetch %s: write: %w", url, err)
	}

	if wantSHA256 != "" {
		got := hex.EncodeToString(h.Sum(nil))
		if got != wantSHA256 {
			os.Remove(tmp.Name())
			return "", fmt.Errorf("fetch %s: sha256 mismatch (got %s want %s)", url, got, wantSHA256)
		}
	}

	return tmp.Name(), nil
}

// Stream returns a ReadCloser for url — useful for large index files
// that we want to parse on the fly without writing to disk.
func Stream(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("stream %s: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("stream %s: HTTP %d", url, resp.StatusCode)
	}
	return resp.Body, nil
}
