$ErrorActionPreference = "Stop"

go mod tidy
go vet ./...
go test ./...
go build -o $null ./cmd/churchtools-invite

Write-Host "CI checks passed."
