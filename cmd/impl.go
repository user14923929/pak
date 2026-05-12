package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/user14923929/pak/internal/cache"
	"github.com/user14923929/pak/internal/dpkg"
	"github.com/user14923929/pak/internal/fetch"
	"github.com/user14923929/pak/internal/index"
	"github.com/user14923929/pak/internal/resolver"
)

// sourcesFile is the repo list, one URL per line (base URL of the repo).
// Example line: https://deb.debian.org/debian bookworm main
const sourcesFile = "/etc/pak/sources.list"

// --- update ---

func runUpdate(_ []string) error {
	sources, err := readSources()
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return fmt.Errorf("no repositories configured — add entries to %s", sourcesFile)
	}

	arch := hostArch()
	for _, src := range sources {
		pkgsURL := src.PackagesURL(arch)
		fmt.Printf(":: fetching %s\n", pkgsURL)

		rc, err := fetch.Stream(pkgsURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: %v\n", err)
			continue
		}

		pkgs, err := index.Parse(rc)
		rc.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: parse error: %v\n", err)
			continue
		}

		if err := cache.Save(pkgsURL, pkgs); err != nil {
			return fmt.Errorf("cache save: %w", err)
		}
		fmt.Printf("   %d packages indexed\n", len(pkgs))
	}
	return nil
}

// --- install ---

func runInstall(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("install: specify at least one package name")
	}

	// Resolve full dependency tree
	plan, err := resolver.Resolve(args, cache.Find)
	if err != nil {
		return err
	}

	fmt.Printf("The following packages will be installed:\n  ")
	names := make([]string, len(plan))
	for i, p := range plan {
		names[i] = p.Name
	}
	fmt.Println(strings.Join(names, "  "))
	fmt.Printf("Proceed? [Y/n] ")

	var ans string
	fmt.Scanln(&ans)
	if ans != "" && strings.ToLower(ans) != "y" {
		fmt.Println("Aborted.")
		return nil
	}

	// Determine base URL (use first source for now)
	sources, err := readSources()
	if err != nil || len(sources) == 0 {
		return fmt.Errorf("no repository configured")
	}
	baseURL := strings.TrimRight(sources[0].URL, "/")

	for _, p := range plan {
		url := baseURL + "/" + p.Filename
		fmt.Printf(":: downloading %s (%s)\n", p.Name, p.Version)

		path, err := fetch.Get(url, p.SHA256)
		if err != nil {
			return err
		}
		defer os.Remove(path)

		fmt.Printf(":: installing  %s\n", p.Name)
		if err := dpkg.Install(path); err != nil {
			return err
		}
	}

	fmt.Println("Done.")
	return nil
}

// --- remove ---

func runRemove(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("remove: specify at least one package name")
	}
	for _, name := range args {
		fmt.Printf(":: removing %s\n", name)
		if err := dpkg.Remove(name); err != nil {
			return err
		}
	}
	fmt.Println("Done.")
	return nil
}

// --- search ---

func runSearch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("search: specify a query")
	}
	query := strings.Join(args, " ")
	pkgs, err := cache.Search(query)
	if err != nil {
		return err
	}
	if len(pkgs) == 0 {
		fmt.Println("No packages found.")
		return nil
	}
	for _, p := range pkgs {
		fmt.Printf("%-30s %s\n    %s\n", p.Name, p.Version, p.Description)
	}
	return nil
}

// --- show ---

func runShow(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("show: specify a package name")
	}
	p, err := cache.Find(args[0])
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("package %q not found", args[0])
	}
	fmt.Printf("Name:         %s\n", p.Name)
	fmt.Printf("Version:      %s\n", p.Version)
	fmt.Printf("Architecture: %s\n", p.Architecture)
	fmt.Printf("Size:         %d bytes\n", p.Size)
	fmt.Printf("SHA256:       %s\n", p.SHA256)
	fmt.Printf("Depends:      %s\n", strings.Join(p.Depends, ", "))
	fmt.Printf("Description:  %s\n", p.Description)
	return nil
}

// --- upgrade ---

func runUpgrade(_ []string) error {
	fmt.Println("upgrade: coming soon — need to compare installed versions vs index")
	return nil
}

// hostArch returns the Debian architecture string for the current machine.
func hostArch() string {
	// runtime.GOARCH → debian arch name
	switch os.Getenv("GOARCH") {
	case "arm64":
		return "arm64"
	case "386":
		return "i386"
	default:
		return "amd64"
	}
}

// --- helpers ---

// Source represents one parsed line from sources.list.
// Format: deb <url> <distro> <component...>
// Example: deb https://deb.debian.org/debian bookworm main contrib
type Source struct {
	URL        string // https://deb.debian.org/debian
	Distro     string // bookworm
	Components []string // [main, contrib]
}

// PackagesURL builds the full URL to the Packages.gz index file.
// e.g. https://deb.debian.org/debian/dists/bookworm/main/binary-amd64/Packages.gz
func (s Source) PackagesURL(arch string) string {
	// use first component only for now
	comp := "main"
	if len(s.Components) > 0 {
		comp = s.Components[0]
	}
	return fmt.Sprintf("%s/dists/%s/%s/binary-%s/Packages.gz",
		strings.TrimRight(s.URL, "/"), s.Distro, comp, arch)
}

func readSources() ([]Source, error) {
	data, err := os.ReadFile(sourcesFile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", sourcesFile, err)
	}

	var out []Source
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// strip leading "deb " or "deb-src " type prefix
		fields := strings.Fields(line)
		if len(fields) < 3 {
			fmt.Fprintf(os.Stderr, "  warning: malformed sources.list line: %q\n", line)
			continue
		}
		start := 0
		if fields[0] == "deb" || fields[0] == "deb-src" {
			start = 1
		}
		if len(fields) < start+2 {
			fmt.Fprintf(os.Stderr, "  warning: malformed sources.list line: %q\n", line)
			continue
		}
		out = append(out, Source{
			URL:        fields[start],
			Distro:     fields[start+1],
			Components: fields[start+2:],
		})
	}
	return out, nil
}
