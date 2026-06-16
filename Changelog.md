# Changelog

All notable changes to this project are documented in this file.

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
