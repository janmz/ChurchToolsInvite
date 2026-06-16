# ChurchTools_Invite

[![Go Version](https://img.shields.io/github/go-mod/go-version/janmz/churchtools-invite)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/janmz/churchtools-invite)](https://github.com/janmz/churchtools-invite/releases)
[![License: MIT (Modified)](https://img.shields.io/badge/License-MIT--Modified-blue.svg)](LICENSE)
[![Support: CFI-Kinderhilfe](https://img.shields.io/badge/Support-CFI--Kinderhilfe-0077B6?logo=heart)](https://cfi-kinderhilfe.de/jetzt-spenden?q=VAYAMASSEN)
[![Build Status](https://github.com/janmz/churchtools-invite/actions/workflows/ci.yml/badge.svg)](https://github.com/janmz/churchtools-invite/actions/workflows/ci.yml)

<p align="center">
  <a href="README.de.md"><img src="https://img.shields.io/badge/🇩🇪-Deutsch-555?style=for-the-badge" alt="Deutsch"></a>
  <img src="https://img.shields.io/badge/🇺🇸-English-0077B6?style=for-the-badge" alt="English (current)">
</p>

**churchtools-invite** is a lightweight Go CLI for **mass ChurchTools system
invitations** from a CSV file, including:

- CSV import with person IDs and optional e-mail sync
- REST API invitations (`POST /persons/{id}/invite`)
- Person export with campus, status and group filters
- Setup, dry-run and permission helpers
- Skip already invited persons by default (`--reinvite` to invite again)

## Features

- Read person IDs from CSV (`id`, `person_id`, `ct_id`, …)
- Send invitation e-mails via the ChurchTools REST API
- Export persons (campus, status, group; interactive selection)
- Setup commands for URL, login token, connection test and permission hints
- Dry-run mode to check CSV and person data before sending
- Sync CSV e-mail to ChurchTools when it differs (old address kept as additional)
- Skip already invited persons by default; use `--reinvite` to invite them again
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

### Go Install

```bash
go install github.com/janmz/churchtools-invite@latest
```

### Build from Source

```bash
git clone https://github.com/janmz/churchtools-invite.git
cd churchtools-invite
go build -o churchtools-invite .
```

On Windows the executable is `churchtools-invite.exe`.

## Usage

### Quick Start

```bash
cp config.example.json config.json
# edit config.json

./churchtools-invite setup test
./churchtools-invite export -o personen.csv
./churchtools-invite invite -f personen.csv --dry-run
./churchtools-invite invite -f personen.csv
```

Global option: `-c config.json` for an alternate config path.

## Configuration

Copy `config.example.json` to `config.json` or use environment variables:

| Variable | Description |
| --- | --- |
| `CT_BASE_URL` | ChurchTools instance URL |
| `CT_LOGIN_TOKEN` | API login token |
| `CT_USERNAME` / `CT_PASSWORD` | Alternative to token |
| `delay_ms` | Delay between invitations in milliseconds (default: 500) |
| `permission_groups.edit_persons` | Group for write access (default: Personen bearbeiten) |
| `permission_groups.export_persons` | Group for export (default: Personen exportieren) |

Obtain a login token:

```bash
./churchtools-invite setup init
# or, after login:
./churchtools-invite setup token
```

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
- Is the person already invited? (by default reported as “skipped”)
- Is there an invitation e-mail (CSV and/or ChurchTools)?
- Would an e-mail sync from the CSV be required?

Output: line-by-line log with `OK`, `SKIPPED` or `ERROR` plus a summary. Exit
code 1 if at least one row failed.

Recommended before the first real run. All invite options (`--reinvite`,
`--no-sync-email`, …) apply to dry-run as well.

## Commands

| Command | Purpose |
| --- | --- |
| `setup init` | Interactive `config.json` creation |
| `setup test` | Test login and connection |
| `setup token` | Show login token |
| `setup permissions` | List invite-related permissions |
| `whoami` | Show logged-in user |
| `export -o FILE` | Export persons to invite CSV format |
| `export -i` | Choose campus and filters interactively |
| `export --campus-id ID` | Export persons from this campus only |
| `export --all-campuses` | No campus filter (default: logged-in user's campus) |
| `export --status-id ID` | Export persons with this status only |
| `export --group-id ID` | Export group members only |
| `invite -f FILE` | Send invitations |
| `invite -f FILE --dry-run` | Check/simulate without sending (see above) |
| `invite -f FILE --no-sync-email` | Skip CSV e-mail sync |
| `invite -f FILE --reinvite` | Invite persons who already have an account again |
| `invite -f FILE --skip-permission-request` | Do not request group membership for missing rights |

## Development

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
original author **Jan Neuhaus** and maintain a link to the original repository:
`https://github.com/janmz/churchtools-invite`.

**No warranty** is provided.

## Support

If you find this project helpful, please support **CFI-Kinderhilfe**:
[https://cfi-kinderhilfe.de](https://cfi-kinderhilfe.de/jetzt-spenden?q=VAYAMASSEN)
(Donations go to CFI-Kinderhilfe, not the author.)

## Contact

**Author**: Jan Neuhaus – [VAYA Consulting](https://vaya-consulting.de/development?q=GITHUB)
**Repository**: [https://github.com/janmz/churchtools-invite](https://github.com/janmz/churchtools-invite)

## Changelog

See [Changelog.md](Changelog.md) for release history.
