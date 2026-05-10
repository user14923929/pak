package cmd

import "fmt"

const usage = `pak — lightweight package manager for dpkg-based systems

Usage:
  pak update              fetch package indexes from repositories
  pak install <pkg...>    install one or more packages
  pak remove  <pkg...>    remove one or more packages
  pak search  <query>     search available packages
  pak show    <pkg>       show package details
  pak upgrade             upgrade all installed packages

Options:
  -h, --help    show this help
`

func Run(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(usage)
		return nil
	}

	sub, rest := args[0], args[1:]

	switch sub {
	case "update":
		return runUpdate(rest)
	case "install":
		return runInstall(rest)
	case "remove":
		return runRemove(rest)
	case "search":
		return runSearch(rest)
	case "show":
		return runShow(rest)
	case "upgrade":
		return runUpgrade(rest)
	default:
		return fmt.Errorf("unknown command %q — run 'pak --help'", sub)
	}
}


