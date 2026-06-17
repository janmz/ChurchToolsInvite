# ChurchTools_Invite

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![Release](https://img.shields.io/badge/Release-GitHub-0077B6)](https://github.com/janmz/ChurchToolsInvite/releases)
[![License: MIT (Modified)](https://img.shields.io/badge/License-MIT--Modified-blue.svg)](LICENSE)
[![Support: CFI-Kinderhilfe](https://img.shields.io/badge/Support-CFI--Kinderhilfe-0077B6?logo=heart)](https://cfi-kinderhilfe.de/jetzt-spenden?q=VAYACTINVITE)
[![Build Status](https://github.com/janmz/ChurchToolsInvite/actions/workflows/ci.yml/badge.svg)](https://github.com/janmz/ChurchToolsInvite/actions/workflows/ci.yml)

<p align="center">
  <a href="README.de.md"><img src="https://img.shields.io/badge/🇩🇪-Deutsch-555?style=for-the-badge" alt="Deutsch"></a>
  <img src="https://img.shields.io/badge/🇺🇸-English-0077B6?style=for-the-badge" alt="English (current)">
</p>

**Churchtools-invite** is a lightweight Go CLI for **mass ChurchTools system
invitations** from a CSV file, including:

- CSV import with person IDs and optional e-mail sync
- REST API invitations (`POST /persons/{id}/invite`)
- Person export with campus, status and group filters
- Setup, dry-run and permission helpers
- Skip already invited persons by default; if the CSV e-mail differs, update
  the address and invite again (`--reinvite` forces invite even when the
  e-mail matches)

## Features

- Read person IDs from CSV (`id`, `person_id`, `ct_id`, …)
- Send invitation e-mails via the ChurchTools REST API
- Export persons (campus, status, group; interactive selection)
- Setup commands for instance name, login token, connection test and permission
  hints
- Dry-run mode to check CSV and person data before sending
- Sync CSV e-mail to ChurchTools when it differs (old address kept as additional)
- Skip already invited persons when the e-mail matches; if the CSV e-mail
  differs, update and invite again; use `--reinvite` to invite all already
  invited persons again
- Automatic group membership request when export or e-mail sync permissions
  are missing

## Requirements

- Go 1.22+ (for building from source)
- ChurchTools account with permission **Invite persons to ChurchTools**
- For export: **export data** permission (group “Personen exportieren”)
- For e-mail sync on invite: **write access** permission (group “Personen
  bearbeiten”)
- Login token or username/password

## Installation

### Binary download

Pre-built binaries for Linux, macOS, and Windows:
[Releases](https://github.com/janmz/ChurchToolsInvite/releases)

Extract the archive and run `churchtools-invite` (`churchtools-invite.exe` on
Windows).

### Go Install

```bash
go install github.com/janmz/churchtools-invite@latest
```

### Build from Source

```bash
git clone https://github.com/janmz/ChurchToolsInvite.git
cd ChurchToolsInvite
go build -o churchtools-invite .
```

On Windows the executable is `churchtools-invite.exe`.

## Usage

### Quick Start

```bash
cp config.example.json config.json
# edit config.json or run setup init

./churchtools-invite setup test
./churchtools-invite export -o personen.csv
```

Edit the list manually, then dry-run:

```bash
./churchtools-invite invite -f personen.csv --dry-run
./churchtools-invite invite -f personen.csv
```

Global option: `-c config.json` for an alternate config path.

## Configuration

Copy `config.example.json` to `config.json` or use environment variables:

| Variable | Description |
| --- | --- |
| `CT_BASE_URL` | Instance name (e.g. `emk-rheinmain`) or full URL |
| `CT_LOGIN_TOKEN` | API login token |
| `CT_USERNAME` / `CT_PASSWORD` | Alternative to token |
| `delay_ms` | Delay between invitations in milliseconds (default: 500) |
| `campus_id` | Default campus when the user has none (set interactively on first export) |
| `permission_groups.edit_persons` | Group for write access (default: Personen bearbeiten) |
| `permission_groups.export_persons` | Group for export (default: Personen exportieren) |

Obtain a login token:

```bash
./churchtools-invite setup init
# or, after login:
./churchtools-invite setup token
```

### Main and sub-instance (OAuth)

In multi-campus ChurchTools setups, a sub-instance URL may look like
`https://main-sub.church.tools` (e.g. `https://emk-rheinmain.church.tools`).
User accounts live on the **central instance**
`https://main.church.tools` (e.g. `https://emk.church.tools`).

If direct login on the sub-instance fails, **username/password** auth runs the
OAuth bridge (when direct login succeeds, OAuth is skipped):

1. Login on the central instance (`/api/login`)
2. `oauthclients/…/startlogin` on the sub-instance (redirect to central)
3. OAuth authorize using the central session
4. Callback on the sub-instance → local session
5. API calls keep using the **sub-instance** (configured URL)

`setup init` can then fetch a sub-instance login token automatically
(`/api/person/me/apitoken`).

If a **login token** is only valid on the central instance, the tool still
falls back to the central URL for API calls (note in output).

Check permissions:

```bash
./churchtools-invite setup permissions
```

## CSV Format

```csv
id,vorname,nachname,email
123,Max,Muster,max@example.org
```

- `id` column is required (also: `person_id`, `ct_id`)
- Name and e-mail columns are optional; e-mail is used to update ChurchTools
  before inviting when it differs from the stored address

### Dry-run – Check before sending

`invite --dry-run` runs the same checks as a real invite but **does not** send
invitations and **does not** change anything in ChurchTools (no e-mail sync).
For each CSV row it verifies:

- Does the person ID exist in ChurchTools?
- Is the person already invited? Detected e.g. via `invitationStatus`
  (`accepted`, `pending`). By default: skip if the CSV e-mail matches
  ChurchTools; if it differs, simulate e-mail update and re-invite
- Is there an invitation e-mail (CSV and/or ChurchTools)?
- Would an e-mail sync from the CSV be required?

Output: line-by-line log with `OK`, `ÜBERSPRUNGEN` or `FEHLER` (German labels)
plus a summary. Exit code 1 if at least one row failed.

Recommended before the first real run. All invite options (`--reinvite`,
`--no-sync-email`, …) apply to dry-run as well.

## Commands

| Command | Purpose |
| --- | --- |
| `setup init` | Interactive `config.json` (instance name, masked password with `*`) |
| `setup test` | Test login and connection |
| `setup token` | Show login token |
| `setup permissions` | List invite-related permissions |
| `whoami` | Show logged-in user, campus ID and effective instance URL |
| `export -o FILE` | Export persons to invite CSV format (default `personen.csv`; `-` = stdout) |
| `export -i` | Choose campus and filters interactively |
| `export --campus-id ID` | Export persons from this campus only |
| `export --all-campuses` | No campus filter (default: user's campus or `campus_id` in config) |
| `export --status-id ID` | Export persons with this status only |
| `export --group-id ID` | Export group members only |
| `export --skip-permission-request` | Do not request group membership for missing export rights |
| `invite -f FILE` | Send invitations |
| `invite -f FILE --dry-run` | Check/simulate without sending (see above) |
| `invite -f FILE --delay-ms MS` | Delay between invitations (0 = `delay_ms` from config) |
| `invite -f FILE --no-sync-email` | Skip CSV e-mail sync (mismatched e-mail → error) |
| `invite -f FILE --reinvite` | Invite persons who already have an account again |
| `invite -f FILE --skip-permission-request` | Do not request group membership for missing rights |

## Development

Linux/macOS release binaries have no embedded file icon (not standard for CLI
tools). Windows release builds embed project-root `vaya.ico` via
[go-winres](https://github.com/tc-hib/go-winres).

**Tests:** Running only `go test` in the repo root does nothing useful (`package
main` has no tests). Use `go test ./...` — see [TESTING.md](TESTING.md) for
coverage, conventions, and manual checks against a real ChurchTools instance.

```bash
go test ./...
go vet ./...
go build -o churchtools-invite .
```

## Contributing

Contributions are welcome! Please check [CONTRIBUTING.md](CONTRIBUTING.md)
before creating a pull request.

## License

This software is under a modified MIT license (see [LICENSE](LICENSE)).
You may freely use, modify, and distribute the code, **provided** you credit the
original author **Jan Neuhaus, VAYA Consulting** and maintain a link to the
original repository: `https://github.com/janmz/ChurchToolsInvite`.

**No warranty** is provided.

## Support

If you find this project helpful, please support **CFI-Kinderhilfe**:
[https://cfi-kinderhilfe.de/jetzt-spenden?q=VAYACTINVITE](https://cfi-kinderhilfe.de/jetzt-spenden?q=VAYACTINVITE)
(Donations go to CFI-Kinderhilfe, not the author.)

## Contact

**Author**: Jan Neuhaus, VAYA Consulting –
[VAYA Consulting](https://vaya-consulting.de/development?q=GITHUB)
**Repository**: [https://github.com/janmz/ChurchToolsInvite](https://github.com/janmz/ChurchToolsInvite)

## Changelog

See [Changelog.md](Changelog.md) for release history.
