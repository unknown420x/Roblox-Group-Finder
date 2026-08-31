# Roblox Group Finder

A minimal, cross-platform Roblox group scanner written in Go.

## Features

- Compatible with all major OS
- Compatible with almost every architecture
- Standard library only (no shitty)
- Persistent configuration
- Checkpoint state support
- Multi-group requests through Roblox's `/v2/groups` endpoint
- Reusable HTTP connections
- Bounded worker concurrency
- Adaptive request limiting
- `429` and `Retry-After` handling
- Optional Discord webhook queue
- Graceful shutdown
- Tests, race detection and `go vet`
- Arch Linux based CI/release containers

## Installation

You can install via releases, community repositories (like AUR) or build from source!


## Build from source

### Requirements

- Git
- Go 1.23+
- GNU Make (Optional)

### Clone

```bash
git clone https://github.com/unknown420x/Roblox-Group-Finder.git
cd Roblox-Group-Finder
```

### Compiling

You can build with GNU Make or with Go compiler

Make:
```bash
make build
```

Go compiler:
```bash
go build -trimpath -ldflags="-s -w" -o groupfinder ./cli
```

## How to run

```bash
./groupfinder
```

If you installed via community repositories run:
```bash
groupfinder
```

Commands:

```text
groupfinder scan
groupfinder config
groupfinder reset
groupfinder version
```

Configuration is saved automatically and reused on the next run.

Example:

```bash
groupfinder scan --workers 4 --rps 1 --batch-size 50
```

## Tests

```bash
go test ./...
go test -race ./...
go vet ./...
```

## Cross compilation

The release workflow builds Linux, Windows and macOS binaries for amd64 and arm64 from an Arch Linux-based CI and release containers.

## API behavior

The scanner uses Roblox's documented multi-get groups endpoint and respects HTTP 429 responses and `Retry-After` guidance. It does not attempt to bypass rate limits.

## License

GNU General Public License v3.0
