package main

import (
	"fmt"
	"os"

	"github.com/yourusername/pak/cmd"
)

// version is injected at build time via -ldflags="-X main.version=..."
var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println("pak", version)
		return
	}
	if err := cmd.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "pak: %v\n", err)
		os.Exit(1)
	}
}
