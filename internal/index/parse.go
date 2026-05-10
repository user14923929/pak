package index

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parse reads an apt Packages (or Packages.gz) stream and returns all packages.
// It handles both plain text and gzip-compressed input transparently.
func Parse(r io.Reader) ([]*Package, error) {
	// detect gzip by trying to wrap; fall back to raw reader
	gr, err := gzip.NewReader(r)
	if err == nil {
		defer gr.Close()
		r = gr
	}

	var pkgs []*Package
	scanner := bufio.NewScanner(r)

	// The format is RFC822-like: stanzas separated by blank lines.
	// Each stanza = one package.
	fields := map[string]string{}

	flush := func() {
		if len(fields) == 0 {
			return
		}
		p := &Package{
			Name:         fields["package"],
			Version:      fields["version"],
			Architecture: fields["architecture"],
			Description:  fields["description"],
			Filename:     fields["filename"],
			SHA256:       fields["sha256"],
		}
		if s, ok := fields["size"]; ok {
			p.Size, _ = strconv.ParseInt(s, 10, 64)
		}
		if dep, ok := fields["depends"]; ok {
			p.Depends = parseDependsList(dep)
		}
		pkgs = append(pkgs, p)
		fields = map[string]string{}
	}

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			flush()
			continue
		}

		// multi-line values start with a space/tab — append to last key
		if line[0] == ' ' || line[0] == '\t' {
			// we don't need multi-line values for now, skip
			continue
		}

		k, v, found := strings.Cut(line, ": ")
		if !found {
			continue
		}
		fields[strings.ToLower(k)] = v
	}
	flush() // last stanza may not end with a blank line

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("index parse: %w", err)
	}
	return pkgs, nil
}

// parseDependsList turns "libc6 (>= 2.17), libssl3" into ["libc6", "libssl3"].
// Version constraints are stripped — good enough for a first pass.
func parseDependsList(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		// alternative deps joined by " | " — take the first alternative only
		part = strings.SplitN(part, " | ", 2)[0]
		// strip version constraint "(>= 2.x)"
		name := strings.Fields(part)[0]
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}
