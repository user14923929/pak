package index

// Package represents one entry from a repository Packages index.
type Package struct {
	Name        string
	Version     string
	Architecture string
	Description string
	Depends     []string // parsed dependency names (simplified)
	Filename    string   // path inside repo, e.g. pool/main/c/curl/curl_7.88.1_amd64.deb
	Size        int64
	SHA256      string
}
