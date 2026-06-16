#!/usr/bin/env bash
set -euo pipefail

go mod tidy
go vet ./...
go test ./...
go build -o /dev/null ./cmd/churchtools-invite

echo "CI checks passed."
