# Masseneinladung

Send ChurchTools system invitations to persons listed in a CSV file.

## Features

- Read person IDs from CSV (`id`, `person_id`, `ct_id`, …)
- Send invitation e-mails via ChurchTools legacy API
  (`invitePersonToSystem`)
- Setup commands for URL, login token, connection test and permission hints
- Dry-run and validate modes before sending
- Sync CSV e-mail to ChurchTools when it differs (old address kept as additional)

## Requirements

- Go 1.22+
- ChurchTools account with permission **Invite persons to ChurchTools**
- Login token or username/password

## Quick start

```bash
go build -o masseneinladung ./cmd/masseneinladung

cp config.example.json config.json
# edit config.json

go run ./cmd/masseneinladung setup test
go run ./cmd/masseneinladung export --output personen.csv
go run ./cmd/masseneinladung validate --csv personen.csv
go run ./cmd/masseneinladung invite --csv personen.csv
```

## Configuration

Copy `config.example.json` to `config.json` or use environment variables:

| Variable | Description |
| --- | --- |
| `CT_BASE_URL` | ChurchTools instance URL |
| `CT_LOGIN_TOKEN` | API login token |
| `CT_USERNAME` / `CT_PASSWORD` | Alternative to token |

Obtain a login token:

```bash
go run ./cmd/masseneinladung setup init
# or, after login:
go run ./cmd/masseneinladung setup token
```

Check permissions:

```bash
go run ./cmd/masseneinladung setup permissions
```

## CSV format

```csv
id,vorname,nachname,email
123,Max,Muster,max@example.org
```

- `id` column is required (also: `person_id`, `ct_id`)
- Name and e-mail columns are optional; e-mail is used to update ChurchTools
  before inviting when it differs from the stored address

## Commands

| Command | Purpose |
| --- | --- |
| `setup init` | Interactive config creation |
| `setup test` | Test login and connection |
| `setup token` | Show login token |
| `setup permissions` | List invite-related permissions |
| `whoami` | Show logged-in user |
| `validate --csv FILE` | Validate CSV without sending |
| `export --output FILE` | Export persons to invite CSV format |
| `export --group-id ID` | Export group members only |
| `invite --csv FILE` | Send invitations |
| `invite --csv FILE --dry-run` | Simulate invitations |
| `invite --csv FILE --no-sync-email` | Skip CSV e-mail sync |

## Development

```bash
go test ./...
go vet ./...
go build .
```

## License

MIT – see [LICENSE](LICENSE).

## Support

Donations for [CFI Kinderhilfe](https://cfi-kinderhilfe.de/?q=VAYAMASSEN).
