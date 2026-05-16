# pak

> A lightweight package manager for dpkg-based Linux systems.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Platform: Linux](https://img.shields.io/badge/platform-Linux-lightgrey)
![Language: Go](https://img.shields.io/badge/language-Go-cyan)

`pak` sits on top of `dpkg` and handles everything it doesn't — fetching
package indexes, resolving dependencies, downloading `.deb` files, and handing
them off to `dpkg` for installation. No daemon, no bloat, no magic.

```
$ pak install curl
The following packages will be installed:
  libssl3  libcurl4  curl
Proceed? [Y/n] y
:: downloading libssl3 (3.0.11-1)
:: installing  libssl3
:: downloading libcurl4 (8.4.0-2)
:: installing  libcurl4
:: downloading curl (8.4.0-2)
:: installing  curl
Done.
```

## Installation

### Download a prebuilt binary

Grab the latest release for your architecture from the
[releases page](../../releases):

```sh
# Linux x86_64
curl -Lo pak https://github.com/user14923929/pak/releases/latest/download/pak-linux-amd64
chmod +x pak
sudo mv pak /usr/local/bin/pak
```

### Build from source

Requires Go 1.22+ or Docker.

**With Go:**
```sh
git clone https://github.com/user14923929/pak
cd pak
go build -o pak .
sudo mv pak /usr/local/bin/pak
```

**With Docker (no Go needed):**
```sh
git clone https://github.com/user14923929/pak
cd pak
make build          # produces ./pak via multi-stage Docker build
sudo mv pak /usr/local/bin/pak
```

## Configuration

Create `/etc/pak/sources.list` with one repository base URL per line.
Lines starting with `#` are ignored.

```sh
sudo mkdir -p /etc/pak
cat <<EOF | sudo tee /etc/pak/sources.list
# Debian bookworm — main archive
https://deb.debian.org/debian/dists/bookworm/main/binary-amd64
EOF
```

## Usage

```
pak update              fetch package indexes from all repositories
pak install <pkg...>    install one or more packages (resolves deps)
pak remove  <pkg...>    remove one or more packages
pak search  <query>     search available packages by name or description
pak show    <pkg>       show detailed info about a package
pak upgrade             upgrade all installed packages  [coming soon]
pak --version           print version
pak --help              print this help
```

## How it works

```
CLI (pak install curl)
        │
        ▼
  Dependency resolver        resolves transitive deps via topological sort
        │
        ▼
  HTTP fetcher               downloads .deb files, verifies SHA256
        │
        ▼
  dpkg -i *.deb              actual installation — pak never touches the fs directly
```

Package indexes are fetched as `Packages.gz` (standard apt format), parsed,
and cached locally under `/var/cache/pak/` as JSON. This means `pak` is
fully compatible with any standard Debian/Ubuntu repository — no custom
server software required.

## Project structure

```
pak/
├── main.go
├── go.mod
├── Dockerfile
├── Makefile
├── cmd/
│   ├── cmd.go          CLI router
│   └── impl.go         subcommand implementations
└── internal/
    ├── index/
    │   ├── package.go  Package type
    │   └── parse.go    Packages.gz parser
    ├── fetch/
    │   └── fetch.go    HTTP downloader + SHA256 verification
    ├── cache/
    │   └── cache.go    local JSON index cache
    ├── resolver/
    │   └── resolver.go dependency resolver (topological sort)
    └── dpkg/
        └── dpkg.go     dpkg wrapper
```

## Roadmap

- [ ] `pak upgrade` — compare installed versions against index
- [ ] `pak list` — list installed packages (via `dpkg-query`)
- [ ] Progress bar during downloads
- [ ] Multiple repository priorities
- [ ] GPG signature verification for indexes

## License

MIT
