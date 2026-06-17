# Changelog

All notable changes to this project are documented in this file.

## [2.3.0.30] - 2026-06-17 10:25:39

### Fixed

- Release-Build (Windows): `embed-windows-icon.sh` ruft `go-winres` über
  `$(go env GOPATH)/bin` auf (CI-PATH enthält dieses Verzeichnis oft nicht)

## [2.2.1.24] - 2026-06-17 10:04:31

### Fixed

- HTTP-401-Behandlung: höchstens ein Re-Login-Versuch pro API-Aufruf (keine
  unbegrenzte Rekursion bei dauerhaftem 401)
- Paginierung: Abbruch nach maximal 10.000 Seiten (Schutz vor Endlosschleifen
  bei fehlerhaften API-Antworten)

## [2.2.1.23] - 2026-06-17 09:47:50

### Added

- `TESTING.md` / `TESTING.de.md`: Teststrategie, `go test ./...`, Abgrenzung zu
  manueller Prüfung gegen echte ChurchTools-Instanzen
- Zusätzliche Unit-Tests: Einladungs-Logik (Live-Einladung, E-Mail-Konflikt,
  Sync bei 403), CLI-Export-Hilfen, Terminal-Passwort (Pipe)

### Changed

- README: Hinweis, warum bloßes `go test` im Root keine Tests ausführt

## [2.2.0.21] - 2026-06-17 09:12:47

### Changed

- Bereits eingeladene Personen werden nur noch übersprungen, wenn die E-Mail aus
  der CSV mit ChurchTools übereinstimmt; bei abweichender Adresse erfolgen
  E-Mail-Update und erneute Einladung (ohne `--reinvite`)

## [2.2.0.20] - 2026-06-17 09:03:23

### Added

- OAuth-Bridge für Nebeninstanzen: Login auf Zentralinstanz, dann
  `oauthclients/…/startlogin` mit Redirect-Folge; API-Session auf der
  konfigurierten Nebeninstanz; `MeAPIToken()` via `/api/person/me/apitoken`
- `setup init` holt nach Passwort-Login bevorzugt den API-Token der Nebeninstanz

### Changed

- Passwort-Login auf Nebeninstanzen nutzt nicht mehr nur die Zentral-URL für
  API-Aufrufe, sondern den vollständigen OAuth-Flow (README aktualisiert)

## [2.1.3.19] - 2026-06-17 08:38:57

### Added

- `setup init`: nur Instanzname (z. B. `emk-rheinmain`) statt voller URL;
  Passwort-Eingabe mit `*`-Maskierung (Windows/Linux/macOS, `golang.org/x/term`)

### Fixed

- `CT_BASE_URL` und `base_url` als Instanzname werden in der vollständige URL
  übersetzt (`Validate` wirkte bisher nicht auf die geladene Config)

## [2.1.3.18] - 2026-06-17 08:33:32

### Fixed

- Hauptinstanz-Fallback auch für Login-Token und CSRF-Abruf (Token gilt oft nur
  auf `haupt.church.tools`, nicht auf `haupt-neben.church.tools`); Session-
  Cookies beim Instanzwechsel nicht mehr verworfen

## [2.1.2.16] - 2026-06-17 08:20:05

### Added

- Login mit Benutzername/Passwort: bei URL-Muster `haupt-neben.church.tools`
  automatischer Versuch auf der Hauptinstanz `haupt.church.tools`; Hinweis bei
  erfolgreichem Wechsel (README: Haupt- und Nebeninstanz)

## [2.1.1.14] - 2026-06-16 22:12:13

### Fixed

- Bereits eingeladene/registrierte Personen werden über `invitationStatus`
  erkannt (`accepted`, `pending`); die bisherigen Felder (`isSystemUser` etc.)
  liefert ChurchTools in Personendetails oft gar nicht mit

## [2.1.1.13] - 2026-06-16 22:06:04

### Fixed

- `invite --dry-run`: bereits eingeladene Personen werden erkannt (ChurchTools
  liefert u. a. `isSystemUser` als Zahl); Ausgabe
  `dry-run: würde überspringen: …`

## [2.1.1.12] - 2026-06-16 21:56:10

### Added

- `whoami`: Standort-ID immer ausgeben (eigene Zeile)
- `config.json`: Feld `campus_id` als Standard-Standort

### Changed

- `export`: ohne `--all-campuses` auf Standort des Nutzers einschränken; fehlt
  dieser, `campus_id` aus config oder einmalige interaktive Auswahl mit
  Speicherung in config

## [2.1.0.10] - 2026-06-16 21:46:08

### Fixed

- Release-Workflow: Artefakt-Upload nutzte ungültiges Glob `dist/*.{tar.gz,zip}`
  (leere Release-Assets); explizite Pfade, Prüfung und `workflow_dispatch` zum
  Nachbauen bestehender Tags

## [2.0.0.8] - 2026-06-16 21:37:33

### Added

- GitHub Actions Release-Workflow: bei Tag `v*` werden Binaries für Linux,
  macOS (amd64/arm64) und Windows gebaut und als Release-Assets veröffentlicht

## [2.0.0.6] - 2026-06-16 20:03:39

### Fixed

- CI: `scripts/ci.sh` baut wieder das Root-Modul (`.`) statt veralteten Pfad
  `./cmd/churchtools-invite`
- README-Badges und Repository-Links auf `janmz/ChurchToolsInvite` korrigiert
  (Go-Version-, Release- und Build-Status-Badge)

## [1.0.6.4] - 2026-06-16 19:43:26

### Removed

- Befehl `validate` entfernt (redundant zu `invite --dry-run`)

### Changed

- README: Prüflauf nur noch über `invite --dry-run` dokumentiert
- Flag-Beschreibung `--dry-run` präzisiert

## [1.0.6.3] - 2026-06-16 19:40:57

### Changed

- Flag `--skip-invited` durch `--reinvite` ersetzt: bereits eingeladene Personen
  werden standardmäßig übersprungen; `--reinvite` lädt erneut ein
- README: ausführliche Erläuterung von `validate` und Vergleich mit
  `invite --dry-run`

## [1.0.6.2] - 2026-06-16 18:44:49

### Fixed

- JSON-Parsing für `privacyPolicyAgreement`: ChurchTools liefert das Feld teils
  als Array statt Objekt (`whoami`, Personendetails); `--skip-invited` schlägt
  damit nicht mehr fehl

### Changed

- README.md und README.de.md: Layout wie wp_plugin_releaser (Badges, Sprachboxen,
  Installation, Lizenz/Kontakt/Changelog-Abschnitte)

## [1.0.6.1] - 2026-06-16 18:07:57

### Added

- Flag `--skip-invited` für `invite` und `validate`: bereits eingeladene
  Personen überspringen (Erkennung über `isSystemUser`, `cmsUserId`,
  `acceptedsecurity`, Datenschutz-Einwilligung)

## [1.0.5.1] - 2026-06-16 17:36:34

### Added

- Automatische Gruppenanfrage beim `export` und `invite` (E-Mail-Sync): fehlende
  Rechte `export data` bzw. `write access` lösen Mitgliedschaftsanfrage für die
  konfigurierten Gruppen aus (Standard: „Personen exportieren“ / „Personen
  bearbeiten“)
- Konfiguration `permission_groups` in `config.json`
- Flag `--skip-permission-request` zum Deaktivieren der automatischen Anfrage

## [1.0.4.1] - 2026-06-16 17:31:39

### Changed

- Einladungen über REST-API `POST /persons/{id}/invite` statt Legacy-AJAX
  (`invitePersonToSystem`)
- Export nutzt standardmäßig den Standort (`campusId`) des angemeldeten Nutzers
- `--all-campuses` deaktiviert den automatischen Standort-Filter

### Fixed

- E-Mail-Sync bei fehlender Berechtigung (403): Hinweis und Einladung trotzdem
  an die ChurchTools-Adresse (statt Abbruch)

## [1.0.3.1] - 2026-06-16 17:27:46

### Added

- Export: Standortauswahl (`--campus-id`) und Filter nach Personenstatus
  (`--status-id`) oder Gruppe (`--group-id`)
- `export --interactive` / `-i`: Standort wählen, danach optional filtern
  (alle Personen, Status oder Gruppe am Standort)
- ChurchTools-API: `/campuses`, `/statuses`, `/groups` mit
  `campus_ids[]`-Filter für Personen

## [1.0.2.2] - 2026-06-16 15:36:03

### Changed

- Alle Go-Pakete auf `package ChurchToolsInvite` umbenannt (Einstiegspunkt
  `cmd/masseneinladung/main.go` bleibt `package main`)
- Import-Aliase für eindeutige Paketreferenzen beibehalten

## [1.0.2.2] - 2026-06-16 15:29:17

### Added

- `export`-Befehl: Personenliste als CSV (`id,vorname,nachname,email`) aus
  ChurchTools exportieren, optional gefiltert nach `--group-id`
- UTF-8-BOM für Excel-kompatible CSV-Dateien

## [1.0.1] - 2026-06-16 15:28:20

### Added

- E-Mail-Sync aus CSV/Excel: abweichende Adresse wird primär gesetzt, bisherige
  ChurchTools-Adresse bleibt als zusätzliche erhalten (PATCH /persons/{id})
- Flag `--no-sync-email` zum Deaktivieren des E-Mail-Syncs

## [1.0.0] - 2026-06-16 15:22:22

### Added

- Initial Go CLI for ChurchTools mass invitations from CSV
- ChurchTools REST and legacy AJAX client with login token / password auth
- Setup commands: `init`, `test`, `token`, `permissions`
- Commands: `invite`, `validate`, `whoami`, `--dry-run`
- Example CSV, config templates, tests and CI workflow
