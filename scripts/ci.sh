#!/usr/bin/env bash
set -euo pipefail

go mod tidy
go vet ./...
go test ./...
go build -o /dev/null ./cmd/masseneinladung

echo "CI checks passed."
