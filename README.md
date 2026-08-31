# Roblox Group finder

Fast, lightweight Roblox group scanner written in Go! :D

## Features

- Fast as shi
- No external Go dependencies
- HTTP connection reuse
- All configurable
- Modular
- Easy use to all skids
- Discord webhook support
- Runtime statistics
- HTTP 429 handling

## Requirements

Go 1.23+

## Build

```bash
go build -ldflags="-s -w" -o groupfinder ./cli
