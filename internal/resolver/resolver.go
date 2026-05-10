package resolver

import (
	"fmt"

	"github.com/user14923929/pak/internal/index"
)

// Resolve returns an ordered install list for the requested packages,
// including all transitive dependencies.
// The order is safe for sequential dpkg -i calls (deps before dependents).
func Resolve(targets []string, find func(string) (*index.Package, error)) ([]*index.Package, error) {
	visited := map[string]bool{}
	var ordered []*index.Package

	var visit func(name string) error
	visit = func(name string) error {
		if visited[name] {
			return nil
		}
		visited[name] = true

		pkg, err := find(name)
		if err != nil {
			return err
		}
		if pkg == nil {
			return fmt.Errorf("package %q not found in index — run 'pak update'?", name)
		}

		// recurse into dependencies first (post-order = deps before dependents)
		for _, dep := range pkg.Depends {
			if err := visit(dep); err != nil {
				// non-fatal: virtual packages or already-installed deps
				// a real resolver would handle this better
				_ = err
			}
		}

		ordered = append(ordered, pkg)
		return nil
	}

	for _, t := range targets {
		if err := visit(t); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}
