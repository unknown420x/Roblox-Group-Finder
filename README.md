# Aleks Group Finder

A minimal, cross-platform Roblox group scanner written in Go.

## Features

- Compatible with all OS
- Compatible with every architecture
- Commons library only (no shitty libraries :D)
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

## Build

```bash
go build -trimpath -ldflags="-s -w" -o groupfinder ./cli
```

## Run

```bash
./groupfinder
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

The release workflow builds Linux, Windows and macOS binaries for amd64 and arm64 from an Arch Linux container.

## API behavior

The scanner uses Roblox's documented multi-get groups endpoint and respects HTTP 429 responses and `Retry-After` guidance. It does not attempt to bypass rate limits.

## License

GNU
