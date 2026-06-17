# Testing

## Quick start

```bash
go test ./...
```

Or run the local CI script (`go vet`, build included):

```bash
scripts/ci.sh     # Linux/macOS
scripts/ci.ps1    # Windows
```

Verbose output:

```bash
go test ./... -v -count=1
```

Coverage summary:

```bash
go test ./... -cover
```

## Why bare `go test` shows “nothing”

The repository root is **`package main`** (entry point only). There are **no**
`*_test.go` files there.

Running `go test` without arguments only tests the **current directory’s
package** — `main` — which reports:

```text
?   github.com/janmz/churchtools-invite   [no test files]
```

All automated tests live under `internal/…` in dedicated test packages (e.g.
`churchtools_test`, `config_test`). You must run **`go test ./...`** (as CI and
the README document).

## What is covered automatically

Tests are **unit/integration tests without a real ChurchTools server**. The REST
API is simulated with `net/http/httptest` (local mock server returning
ChurchTools-shaped JSON).

| Area | Package / tests | Examples |
| --- | --- | --- |
| API client | `internal/churchtools`, `internal/churchtools_test` | Login, CSRF, persons, email update, invite, campuses/groups, permissions, pagination |
| OAuth / sub-instances | `internal/churchtools` | Central login, redirect chain, sub-instance session |
| Person JSON | `internal/churchtools` | `invitationStatus`, legacy fields, privacy consent |
| Invite logic | `internal/invite` | Dry-run, skip rules, email mismatch, sync, live invite |
| Configuration | `internal/config_test` | Load/save, env overrides, validation |
| CSV | `internal/csvfile_test` | Read/write, columns, roundtrip |
| CLI helpers | `internal/cmd` (partial) | Export filter labels, campus names |
| Terminal | `internal/termio` | Password from pipe (non-TTY) |

Currently **40+** test functions across **14** files.

## What is intentionally not fully automated

| Area | Reason |
| --- | --- |
| **`cmd/` (Cobra commands)** | Thin wrapper over `internal/*`; interactive flows (`setup init`, campus menus) need a TTY. Business logic is tested in `internal/`. |
| **Real ChurchTools server** | No shared test instance in CI; permissions and data differ per church. Manual checks: `churchtools-invite setup test` and `invite --dry-run`. |
| **Email delivery / SMTP** | Invitations are sent by ChurchTools; this tool only calls `POST /persons/{id}/invite` (covered by mocks). |
| **Interactive password (TTY)** | Raw mode and `*` echo are platform-specific; piped input is tested. |
| **`main` / version constants** | Nothing meaningful to unit-test. |

### Manual acceptance on a real instance

1. Copy `config.example.json` to `config.json` (do not commit).
2. `churchtools-invite setup test` — login and connectivity.
3. `churchtools-invite export` — export a small set of persons.
4. `churchtools-invite invite -f personen.csv --dry-run` — verify without sending.
5. Only then run without `--dry-run` (preferably with test persons).

**Do not use production personal data** in tests or commits.

## Conventions

- **External test packages** (`churchtools_test`, …) exercise the public API.
- **Internal tests** (`package churchtools`) cover unexported helpers.
- Mocks: `httptest.NewServer` with handlers for `/api/whoami`, `/api/csrftoken`,
  `/api/persons/…`, etc.
- Filesystem: `t.TempDir()`; environment: `t.Setenv`.

## Run a single package

```bash
go test ./internal/invite/... -v
go test ./internal/churchtools/... ./internal/churchtools_test/... -v
```

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs `scripts/ci.sh`, which
includes `go test ./...`.
