# Tests

## Schnellstart

```bash
go test ./...
```

Oder das lokale CI-Skript (inkl. `go vet`, Build):

```bash
scripts/ci.ps1    # Windows
scripts/ci.sh     # Linux/macOS
```

Mit Ausgabe pro Testfall:

```bash
go test ./... -v -count=1
```

Mit Abdeckungsübersicht:

```bash
go test ./... -cover
```

## Warum `go test` ohne `./...` „nichts“ zeigt

Im Projektroot liegt nur `package main` (Einstiegspunkt). Dort gibt es **keine**
`*_test.go`-Dateien.

Ein bloßes `go test` prüft nur das Paket im aktuellen Verzeichnis — also
`main` — und meldet:

```text
?   github.com/janmz/churchtools-invite   [no test files]
```

Alle automatisierten Tests liegen unter `internal/…` in eigenen Testpaketen
(z. B. `churchtools_test`, `config_test`). Deshalb ist **`go test ./...`**
Pflicht (steht auch in CI und README).

## Was automatisiert getestet wird

Die Tests sind **Unit- und Integrationstests ohne echten ChurchTools-Server**.
Stattdessen simulieren sie die REST-API mit `net/http/httptest` (lokaler
Mock-Server mit ChurchTools-ähnlichen JSON-Antworten).

| Bereich | Paket / Tests | Inhalt (Auszug) |
| --- | --- | --- |
| API-Client | `internal/churchtools`, `internal/churchtools_test` | Login, CSRF, Personen laden, E-Mail-Update, Einladung, Campus/Gruppen, Berechtigungen, Paginierung |
| OAuth / Nebeninstanzen | `internal/churchtools` | Zentral-Login, Redirect-Kette, Sub-Instanz-Session |
| Person-JSON | `internal/churchtools` | `invitationStatus`, Legacy-Felder, Datenschutz-Einwilligung |
| Einladungs-Logik | `internal/invite` | Dry-Run, Überspringen, E-Mail-Abweichung, Sync, Live-Einladung |
| Konfiguration | `internal/config_test` | Laden/Speichern, Umgebungsvariablen, Validierung |
| CSV | `internal/csvfile_test` | Lesen/Schreiben, Spalten, Roundtrip |
| CLI-Hilfslogik | `internal/cmd` (teilweise) | Export-Filter-Beschreibung, Campus-Namen |
| Terminal | `internal/termio` | Passwort aus Pipe (nicht-TTY) |

Stand: über **40** Testfunktionen in **14** Dateien (Zahl kann wachsen).

## Was bewusst nicht vollständig automatisiert ist

| Bereich | Grund |
| --- | --- |
| **`cmd/` (Cobra-Befehle)** | Dünne Schicht über `internal/*`; interaktive Menüs (`setup init`, Campus-Auswahl) brauchen TTY und manuelle Eingaben. Die Geschäftslogik wird in `internal/` getestet. |
| **Echter ChurchTools-Server** | Keine feste Test-Instanz in CI; API-Details und Rechte variieren pro Gemeinde. Manuelle Prüfung: `churchtools-invite setup test` und `invite --dry-run`. |
| **E-Mail-Versand / SMTP** | Einladungen löst ChurchTools serverseitig aus; das Tool ruft nur `POST /persons/{id}/invite` auf (im Mock getestet). |
| **Interaktive Passwort-Eingabe (TTY)** | Raw-Modus und `*`‑Echo sind plattformabhängig; Pipe-Eingabe wird getestet. |
| **`main` / Versionskonstanten** | Kein sinnvoller Unit-Test. |

### Manuelle Abnahme gegen eine echte Instanz

1. `config.json` aus `config.example.json` (nicht committen).
2. `churchtools-invite setup test` — Login und Verbindung.
3. `churchtools-invite export` — kleine Personenmenge exportieren.
4. `churchtools-invite invite -f personen.csv --dry-run` — ohne Versand prüfen.
5. Erst danach ohne `--dry-run` einladen (am besten mit Testpersonen).

**Keine personenbezogenen Produktivdaten** in Tests oder Commits verwenden.

## Test-Konventionen

- **Externe Testpakete** (`churchtools_test`, `config_test`, …) testen die
  öffentliche API wie ein Aufrufer.
- **Interne Tests** (`package churchtools`) für unexportierte Helfer (JSON-Decode,
  OAuth-Discovery).
- Mock-Server: `httptest.NewServer` + Handler für `/api/whoami`, `/api/csrftoken`,
  `/api/persons/…` usw.
- Dateisystem: `t.TempDir()`; Umgebung: `t.Setenv`.

## Einzelnes Paket testen

```bash
go test ./internal/invite/... -v
go test ./internal/churchtools/... ./internal/churchtools_test/... -v
```

## CI

GitHub Actions (`.github/workflows/ci.yml`) führt `scripts/ci.sh` aus —
dort ebenfalls `go test ./...`.
