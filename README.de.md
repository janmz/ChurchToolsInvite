# Masseneinladung

Versendet ChurchTools-Systemeinladungen an Personen aus einer CSV-Liste.

## Funktionen

- CSV einlesen mit Personen-IDs (`id`, `person_id`, `ct_id`, …)
- Einladungs-E-Mails über die ChurchTools-API senden
  (`invitePersonToSystem`)
- Setup-Befehle für URL, Login-Token, Verbindungstest und Berechtigungshinweise
- Dry-Run und Validierung vor dem Versand
- E-Mail aus CSV/Excel übernehmen: abweichende Adresse wird primär gesetzt,
  bisherige ChurchTools-Adresse bleibt als zusätzliche erhalten

## Voraussetzungen

- Go 1.22+
- ChurchTools-Konto mit Berechtigung **Personen zur Nutzung von ChurchTools einladen**
- Login-Token oder Benutzername/Passwort

## Schnellstart

```bash
go build -o masseneinladung.exe ./cmd/masseneinladung

copy config.example.json config.json
# config.json anpassen

go run ./cmd/masseneinladung setup test
go run ./cmd/masseneinladung export --output personen.csv
go run ./cmd/masseneinladung validate --csv personen.csv
go run ./cmd/masseneinladung invite --csv personen.csv
```

## Konfiguration

Kopiere `config.example.json` nach `config.json` oder nutze Umgebungsvariablen:

| Variable | Beschreibung |
| --- | --- |
| `CT_BASE_URL` | URL der ChurchTools-Instanz |
| `CT_LOGIN_TOKEN` | API-Login-Token |
| `CT_USERNAME` / `CT_PASSWORD` | Alternative zum Token |

Login-Token beschaffen:

```bash
go run ./cmd/masseneinladung setup init
# oder nach Login:
go run ./cmd/masseneinladung setup token
```

Berechtigungen prüfen:

```bash
go run ./cmd/masseneinladung setup permissions
```

## CSV-Format

```csv
id,vorname,nachname,email
123,Max,Muster,max@example.org
```

- Spalte `id` ist Pflicht (auch: `person_id`, `ct_id`)
- Name und E-Mail sind optional; bei abweichender E-Mail wird ChurchTools vor
  dem Einladen aktualisiert (alte Adresse bleibt als zusätzliche erhalten)

## Befehle

| Befehl | Zweck |
| --- | --- |
| `setup init` | Interaktive config.json anlegen |
| `setup test` | Login und Verbindung testen |
| `setup token` | Login-Token anzeigen |
| `setup permissions` | Einladungs-Berechtigungen prüfen |
| `whoami` | Angemeldeten Benutzer anzeigen |
| `validate --csv DATEI` | CSV prüfen ohne Versand |
| `export --output DATEI` | Personenliste als Einladungs-CSV exportieren |
| `export --group-id ID` | Nur Gruppenmitglieder exportieren |
| `invite --csv DATEI` | Einladungen senden |
| `invite --csv DATEI --dry-run` | Einladungen simulieren |
| `invite --csv DATEI --no-sync-email` | E-Mail-Sync aus CSV deaktivieren |

## Entwicklung

```bash
go test ./...
go vet ./...
go build .
```

## Lizenz

MIT – siehe [LICENSE](LICENSE).

## Unterstützung

Spenden für die [CFI Kinderhilfe](https://cfi-kinderhilfe.de/?q=VAYAMASSEN).
