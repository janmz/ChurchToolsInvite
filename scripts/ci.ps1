$ErrorActionPreference = "Stop"

go mod tidy
go vet ./...
go test ./...
go build -o $null ./cmd/masseneinladung

Write-Host "CI checks passed."
