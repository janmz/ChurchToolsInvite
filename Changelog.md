# Changelog

All notable changes to this project are documented in this file.

## [2.8.0.56] - 2026-11-10 17:22:27

### Added

- Neuer Befehl `list` mit `list statuses`, `list groups [--campus X]
  [--search TEXT]` und `list campuses`: zeigt ID und Name der Werte für
  `export --status`, `--group` und `--campus` (jede Gemeinde legt eigene
  Status und Gruppen an, die IDs unterscheiden sich je Instanz)

### Changed

- `export --status` und `export --group` akzeptieren ID **oder** Namen
  (exakter Name vor eindeutigem Teilstring, Umlaute/Groß-/Kleinschreibung
  egal), z. B. `--status Mitglied`; `--status-id`/`--group-id` funktionieren
  weiter als Alias
- Unbekannte IDs, nicht gefundene oder mehrdeutige Namen werden vor dem
  Export abgelehnt; die Meldung nennt die gültigen Werte (bei vielen
  Einträgen den passenden `list`-Befehl). Das gilt jetzt auch für
  `--campus` mit ID
- Hilfetext und README unterscheiden Personenstatus (`--status`) und
  Einladungsstatus (CSV-Spalte `status`)

## [2.7.2.55] - 2026-11-10 17:09:11

Umsetzung des Sicherheitsaudits vom 2026-11-10 (Nr. 1 – 15).

### Security

- `invite` vergleicht den Namen aus der CSV mit der Person in ChurchTools
  (Umlaute/Akzente/Groß-/Kleinschreibung egal); bei Abweichung schlägt die
  Zeile fehl, außer mit `--ignore-name-mismatch` – eine verrutschte Zeile
  schickt sonst die Kontoaktivierung an eine fremde Adresse (Nr. 2)
- Vor einem echten `invite`-Lauf wird ein Plan erstellt (Zahl der
  Einladungen, jede E-Mail-Änderung alt → neu, fehlerhafte Zeilen) und mit
  `[j/N]` bestätigt; `--yes`/`-y` für Skripte (Nr. 2)
- Temporär beigetretene Gruppen werden in `.churchtools-invite-groups.json`
  (neben `config.json`, nur für den Benutzer lesbar) vermerkt; nach Absturz,
  geschlossenem Fenster oder zweitem Strg+C bietet der nächste Start den
  Austritt an. Nicht verlassene Gruppen bleiben vermerkt (Nr. 3)
- Temporärdateien beim atomaren Schreiben werden vor dem ersten Byte auf den
  Benutzer beschränkt; `.*.tmp-*` in `.gitignore` (Nr. 4)
- E-Mail-Adressen aus der CSV werden geprüft (genau eine einfache Adresse,
  max. 254 Zeichen, keine Leer-/Steuerzeichen); das Ersetzen der primären
  Adresse bei abgelehnter Adressliste nur noch mit `--allow-single-email`
  (Nr. 5)
- CSV-Dateien werden beim Lesen auf den aktuellen Benutzer beschränkt
  (Warnung, wenn nicht möglich) (Nr. 1)
- Token-Fallback auf `main_url` nur noch bei HTTP 401; läuft die API über
  die Hauptinstanz, warnt `invite` und verlangt eine interaktive Bestätigung
  (`--yes` wird abgelehnt) (Nr. 6)
- Ist der Mitgliedschaftsstatus nicht lesbar, wird vor dem Beitritt die
  Gruppenliste der Person geprüft – eine bestehende Mitgliedschaft wird nie
  als neu gewertet und am Ende entfernt (Nr. 7)
- `setup init` schlägt die Hauptinstanz vor, an die die Nebeninstanz ihren
  OAuth-Login selbst weiterleitet; ein aus dem Namen geratener Host muss
  eingetippt werden (`j` gilt dafür nicht) (Nr. 8)
- OAuth-Brücke: Weiterleitungen nur noch zu Haupt- und Nebeninstanz und nur
  per https (http nur localhost); die Nebeninstanz muss genau die
  konfigurierte Hauptinstanz verwenden; Antworten größenbegrenzt (Nr. 9)
- Namen und Texte von Server und CSV werden ohne Steuerzeichen und
  Bidi-Steuerzeichen ausgegeben (Nr. 10)
- Ergebniszeilen eines echten Laufs maskieren E-Mail-Adressen; vollständige
  Adressen nur im Dry-Run und im Plan (Nr. 11)
- `setup init` fragt, bevor ein dauerhafter Login-Token abgerufen bzw.
  erzeugt wird (Nr. 12)
- Passwort-/Token-Eingabe ohne Konsole (z. B. Git Bash/mintty) zeigt einen
  Hinweis, dass die Eingabe sichtbar sein kann (Nr. 13)

### Changed

- Go-Mindestversion 1.26 (Go 1.25 erhält keine Sicherheitsupdates mehr);
  CI-Testjob mit Go 1.26 (Nr. 14)
- Release: Build ohne Actions-Cache, `SHA256SUMS` für alle Archive (Nr. 15)
- Pre-Commit-Hook `.githooks/pre-commit` (aktivieren mit
  `git config core.hooksPath .githooks`): CI-Skript, EN/DE-Dokupaare,
  Sperre für CSV, `config.json` und Prüfberichte (Nr. 15)

### Dependencies

- `golang.org/x/sys` v0.29.0 → v0.49.0, `golang.org/x/term` v0.28.0 →
  v0.47.0, `github.com/spf13/cobra` v1.9.1 → v1.10.2 (`spf13/pflag` v1.0.9);
  alle mindestens 14 Tage veröffentlicht (Nr. 14)

### Tests

- Neue Tests für OAuth-Discovery und abgelehnte Fremd-Weiterleitungen,
  Hauptinstanz-Vorschlag, Zustandsdatei der Gruppen, ACL der Temp-Datei,
  E-Mail-Prüfung, verweigerten Einzeladress-Fallback, Steuerzeichenfilter

## [2.7.1.54] - 2026-11-10 16:32:38

Umsetzung des Code-Reviews vom 2026-11-10 (CR-1 … CR-24).

### Fixed

- `setup init`/`setup token`: Beim Wechsel zwischen Token und Passwort wird
  auch die verschlüsselte Kopie der alten Zugangsdaten entfernt; bisher war
  nach dem Speichern wieder der alte Token aktiv (CR-1)
- `setup init`: Die Antwort `n`/`nein`/`-` auf die Frage nach der
  Hauptinstanz bedeutet „keine“ (vorher wurde daraus
  `https://n.church.tools`); Eingaben werden geprüft, `main_url` gleich
  `base_url` wird abgelehnt (CR-2)
- `export -o - --interactive`: Status- und Hinweiszeilen gehen nach stderr,
  die CSV auf stdout bleibt sauber (CR-3)
- Strg+C bricht nach der aktuellen Zeile bzw. Eingabe ab; temporär
  beigetretene Gruppen werden dabei wieder verlassen, die Zusammenfassung
  wird ausgegeben; ein zweites Strg+C beendet sofort (CR-4)
- E-Mail-Abgleich: Ist keine primäre Adresse gesetzt, bleiben die übrigen
  Adressen der Person erhalten (CR-5)
- `--no-sync-email`: Personen ohne E-Mail in ChurchTools werden schon im
  Dry-Run als Fehler gemeldet (ChurchTools lädt nur an die eigene Adresse
  ein) (CR-6)
- Ausweich-PATCH mit nur einer Adresse nur noch bei HTTP 400/422; die Meldung
  sagt dann korrekt, dass die bisherige Adresse ersetzt wurde (CR-9)
- Instanzen ohne Standorte: kein Menüfehler mehr; ohne Eingabemöglichkeit
  Hinweis auf `--campus` (CR-10)
- Fehlende Unterstützung für Zugriffsrechte (FAT, Netzlaufwerk) führt nur
  noch zu einer Warnung statt jeden Befehl abzubrechen (CR-11)
- Fehlermeldungen enthalten den Grund vom Server (z. B. warum eine Einladung
  abgelehnt wurde) (CR-12)
- Rechtegruppen: mehrdeutige Gruppennamen und Netzfehler werden gemeldet
  statt „keine passende Gruppe gefunden“ (CR-13)
- Vorab-Gruppen: beantragte oder zum Löschen markierte Mitgliedschaften
  erscheinen nicht mehr als „bereits Mitglied“ (CR-14)
- Echter Lauf meldet „übersprungen“ statt „würde überspringen“ (CR-15)
- Standortanzeige ohne doppelte ID („ID 7 (ID 7)“), Menüpunkt „Alle
  Standorte“ ohne „ID -1“ (CR-16)
- `config.json` und CSV werden atomar geschrieben (Temp-Datei + Umbenennen);
  ein Abbruch hinterlässt keine leere Datei mehr (CR-17)

### Changed

- Dry-Run: Vorab- und Rechtegruppen werden weiterhin beigetreten (Leserechte),
  am Ende aber – wie bei jedem Lauf – wieder verlassen, wenn vorher keine
  Mitgliedschaft bestand (CR-7, nach Rückfrage abgewandelt)
- `invite` gibt jedes Ergebnis sofort aus; die Zusammenfassung wiederholt
  nur die Fehler (CR-8)
- `delay_ms`/`--delay-ms`: nur 0–60000 ms; gewartet wird nur nach einer
  Einladung oder E-Mail-Änderung, nicht nach übersprungenen Zeilen (CR-18,
  CR-19)
- `invite` lädt die Personen der CSV blockweise (50 je Anfrage) statt
  einzeln; Standorte und Gruppen des Benutzers werden je Lauf bzw. Durchgang
  nur einmal geladen (CR-20)
- `setup init` schreibt keine Standardwerte (`delay_ms`, `pre_join_groups`)
  mehr fest in die config; leere Eingabe behält den bisherigen Wert (CR-23)
- CI prüft `go.mod`/`go.sum` mit `go mod tidy -diff`, statt sie still zu
  reparieren (CR-24)

### Removed

- Ungenutzter Code: `maskEmail`, `allowSkip` in Menüs, `PermInvitePersons`,
  `ModuleChurchCore`, `Client.HasModulePermission`, ungenutzte
  `cfg`-Parameter; irreführende Kommentare und Testnamen aus dem dup-Merge
  korrigiert (CR-21, CR-22)

### Tests

- Neue Tests für Login-Wechsel mit verschlüsselten Feldern, Hauptinstanz-
  Eingaben, Standardwerte, `delay_ms`-Grenzen, atomares Schreiben,
  E-Mail-Plan ohne Primäradresse, Fallback nur bei 400/422,
  Server-Fehlertext, mehrdeutige Gruppen, Vorabladen der Personen, Pausen nur
  nach Schreibzugriffen, sofortige Ausgabe und Abbruch

## [2.7.0.53] - 2026-11-10 12:04:13

Übernahme der Optimierungen aus `churchtools-dup` (Code-Review und
Sicherheitsaudit vom 2026-11-10, Versionen 1.3.0.34 – 1.3.2.36).

### Security

- Keine Zugangsdaten mehr an abgeleitete Hosts: Der Login-Fallback auf eine
  „Hauptinstanz“ wird nur noch mit ausdrücklich gesetztem `main_url`
  (`CT_MAIN_URL`) genutzt; `setup init` schlägt sie vor und übernimmt sie nur
  nach Bestätigung
- `login_token` und `password` werden mit `github.com/janmz/sconfig/v3`
  rechnergebunden verschlüsselt in `config.json` gespeichert; vorhandene
  Klartextwerte werden beim ersten Start umgewandelt
- `config.json` und exportierte CSV-Dateien nur für den aktuellen Benutzer
  lesbar (Windows-ACL ohne Vererbung, sonst `0600`)
- Gruppen nur noch über den exakten Namen (Sortiernummern-Markup wird
  ignoriert); öffentliche Gruppenanmeldung nur nach Rückfrage
- CSV-Export entschärft Zellen mit `= + - @`/Tab/CR (Formel-Injection);
  `invite` stellt die Werte wieder her
- `setup token`: Option `--person-id` entfernt (keine Tokens fremder Konten);
  der eigene Token wird nicht mehr ausgegeben, sondern nach Rückfrage
  verschlüsselt gespeichert
- Nur `https://` für `base_url`/`main_url` (`http://` nur localhost);
  Weiterleitungen auf anderen Host oder anderes Schema werden abgelehnt
- Login-Token wird verdeckt eingegeben; Passworteingabe ohne Sternchen
- Servertexte in Meldungen gekürzt und ohne Steuerzeichen

### Fixed

- Passwort-Sitzungen: Einladung (`POST /persons/{id}/invite`), E-Mail-Sync
  (`PATCH /persons/{id}`) und Gruppenbeitritt senden jetzt das CSRF-Token im
  Header `CSRF-Token`; neuere ChurchTools-Versionen lehnen Schreibzugriffe aus
  einer Sitzung sonst ab
- Alle API-Aufrufe laufen über eine zentrale HTTP-Schicht: Größenlimit je
  Antwort (32 MiB) und je Liste (512 MiB), eine Wiederholung bei HTTP 429,
  erneute Anmeldung nach Sitzungsablauf auch bei Benutzername/Passwort
- Gruppenbeitritt: bestehende Mitgliedschaften werden nicht mehr per PUT
  überschrieben (Kommentar/Enddatum blieben sonst nicht erhalten); neuer
  Beitritt mit `only_add`
- Passworteingabe: Umlaute bleiben erhalten, Leerzeichen am Rand werden nicht
  mehr abgeschnitten (Terminal, Pipe, `CT_PASSWORD`)
- Setup über Pipe: alle Eingaben aus demselben Eingabepuffer; letzte Zeile ohne
  Zeilenumbruch wird akzeptiert; Menüs und Hinweise gehen nach stderr
  (`export -o -` liefert reine CSV)
- `setup init` fragt vor dem Überschreiben und behält `campus_id`;
  Standortauswahl speichert nur `campus_id` (keine Werte aus
  Umgebungsvariablen in `config.json`)
- `--campus`: Umlaute/Akzente werden umgeschrieben (ü→ue, é→e), Ziffern
  bleiben erhalten, exakter Name hat Vorrang

### Changed

- Für einen Lauf beigetretene Vorab- und Rechtegruppen werden am Ende von
  `export`/`invite` wieder verlassen (auch bei Fehlern); bestehende
  Mitgliedschaften bleiben unverändert
- `export` überschreibt eine vorhandene Ziel-Datei nur noch mit `--force`
- `invite` lehnt eine CSV mit doppelter Personen-ID vor jedem Versand ab
  (mit Zeilennummern); CSV-Dateien über 100 MiB werden abgelehnt
- Datenschutzhinweis zum Löschen der CSV nach Export und Einladung; neuer
  README-Abschnitt „Zugangsdaten und Datenschutz“
- Go-Mindestversion 1.25 (sconfig); CI: `gofmt`-Prüfung, `staticcheck` v0.8.2
  (eigener Lint-Job mit Go 1.27), Standardrechte `contents: read`, Actions per
  Commit-SHA fixiert; `.gitattributes` erzwingt LF für Go-Dateien

### Removed

- Versehentlich eingecheckte Binärdateien `cmd/churchtools-invite` und
  `churchtools-invite-test.exe` aus dem Repository (`*.exe` wird ignoriert)
- Ungenutzter Code (`refreshSession`, `containsEmail`, `newCSVWriter`)

### Dependencies

- `github.com/janmz/sconfig/v3` v3.0.0 (wie in `churchtools-dup`, Ausnahme von
  der 14-Tage-Regel dort nach Rückfrage), `golang.org/x/sys` v0.29.0 jetzt direkt

### Tests

- Neue Tests für zentrale HTTP-Schicht (HTTP 429, Re-Login per Passwort,
  CSRF-Header bei Einladung/PATCH/PUT), exakte Gruppensuche, Konfigurations-
  verschlüsselung, Dateirechte, Größenlimits, Textnormalisierung,
  CSV-Formel-Roundtrip und doppelte Personen-IDs

## [2.6.3.52] - 2026-08-19 22:39:42

### Removed

- Personenlisten aus dem Repository entfernt (`20260705-Liste.csv/.xlsx`,
  `dup20260706.csv`, `stand19072026.csv`) – auch aus der gesamten Git-Historie

### Changed

- `.gitignore`: generell `*.csv` und `*.xlsx` ignorieren (Ausnahme:
  `examples/*.example`)

## [2.6.2.50] - 2026-08-19 22:28:17

### Changed

- CLI-Hilfe und Fehlermeldungen vollständig auf Deutsch (Verwendung, Optionen,
  Pflichtoptionen, unbekannte Befehle/Optionen)
- Einheitliche Fehlerausgabe: genau eine Meldung, bei Parsing- und
  Validierungsfehlern immer der passende Hilfe-Block
- `setup` ohne oder mit unbekanntem Unterbefehl beendet mit Fehlercode 1
- `invite`: ausführlicher Hilfetext wie bei `export`

### Fixed

- Doppelte Fehlermeldungen bei Cobra-Fehlern (zuvor Error-Zeile plus Wiederholung
  in `main`)
- Runtime-Fehler (`--csv`/`--output` fehlt, Pfad sieht wie Option aus) zeigen
  jetzt ebenfalls die Befehls-Hilfe

## [2.6.1.48] - 2026-06-20 12:47:14

### Added

- `export --interactive`: Einladungsstatus wählen mit `[n]` Neu, `[e]` Eingeladen,
  `[r]` Registriert

### Changed

- `README.md` / `README.de.md`: CSV-Spalte `standort`, `--reinvite`-Verhalten,
  interaktiver Einladungsstatus, `whoami`-Gruppenformat

## [2.6.1.47] - 2026-06-20 12:40:33

### Fixed

- Gruppenbeitritt per Signup (`autoAccept`): Rückgabe enthält jetzt die Meldung
  „Mitgliedschaft angenommen (Gruppenanmeldung)“ statt leerer Message bei Status
  `active`

## [2.6.1.46] - 2026-06-20 12:39:17

### Fixed

- `invite`: Personen mit Status „Registriert“ werden nie erneut eingeladen,
  auch nicht mit `--reinvite` (z. B. nach Export mit `--invited`)

### Changed

- `--reinvite` gilt nur noch für „Eingeladen“ (Einladung ausstehend)

## [2.6.1.45] - 2026-06-20 12:36:42

### Fixed

- `export --output` / `invite --csv`: Werte wie `-i` oder `--invited` nach `-o`/`-f`
  werden als fehlender Dateiname erkannt statt still als Dateiname interpretiert

## [2.6.1.44] - 2026-06-20 12:27:54

### Fixed

- Gruppenbeitritt: Bei `403` auf `PUT /groups/{id}/members/{personId}` wird die
  Web-GUI-Anmeldung über `GET /publicgroups/{id}/form`, `POST .../token` und
  `POST .../signup` verwendet (Behebt „denied“ trotz sichtbarem „Beitreten“-Button)

## [2.6.1.43] - 2026-06-20 12:14:10

### Fixed

- `whoami` / `ListPersonGroups`: `groupId` statt Mitgliedschafts-`id` verwenden
  (behebt doppelte/falsche Gruppen-ID 932 o. ä.)
- `whoami`: Gruppen alphabetisch nach Name, Ausgabe `Name  ID`

## [2.6.1.42] - 2026-06-20 11:55:55

### Added

- `whoami`: Gruppenmitgliedschaften des angemeldeten Benutzers mit ID und Name

## [2.6.1.41] - 2026-06-20 11:12:23

### Fixed

- `pre_join_groups`: noch unsichtbare Gruppen blockieren nicht mehr die restliche
  Liste; nach erfolgreichem Beitritt wird die Sitzung erneuert und vorher
  nicht gefundene Gruppen erneut versucht

### Changed

- Standard `pre_join_groups`: `ChurchTools Verwaltung,Gruppen Administration,ChurchTools Admin,Personen Administration,Personen verwalten`

## [2.6.0.40] - 2026-06-20

Feature: Now only uninvited persons are exported unless `--invited` is given. `--campus-id` becomes `--campus` and accepts (partial) names, csv includes campus and status

## [2.5.0.39] - 2026-06-20 08:59:33

### Added

- Export-Spalte `standort` (nach `email`) mit Standortnamen aus ChurchTools

### Removed

- Konfiguration `permission_groups` entfernt; fehlende Export-/Schreib-Rechte
  werden weiterhin automatisch über feste Fallback-Gruppen beantragt
  (`Personen exportieren` bzw. `Personen Administration` / `Personen bearbeiten`),
  Vorab-Beitritt über `pre_join_groups`

## [2.5.0.38] - 2026-06-20 08:56:05

### Added

- `export --invited` / `-i`: auch bereits eingeladene und registrierte Personen
  exportieren
- Export-Spalte `status` mit `NEU`, `Eingeladen`, `Registriert`

### Changed

- `export`: standardmäßig nur Personen mit Status `NEU` (noch nicht eingeladen)
- `export --campus-id` durch `--campus` ersetzt: numerische ID, `all` oder
  eindeutiger Namens-Teilstring (Kleinbuchstaben, Suchstring nur `a`–`z`)
- `export --interactive` ohne Kurzoption `-i` (`-i` ist jetzt `--invited`)

## [2.5.0.37] - 2026-06-19 17:51:06

### Added

- Vorab-Gruppen (`pre_join_groups`): vor Export/Invite nacheinander Gruppen
  beitreten, wenn noch nicht Mitglied; `setup init`-Abfrage; `--skip-pre-join-groups`
- `export --campus-id all` (auch `alle`, `*`); Alias zu `--all-campuses`
- CSV-Import: UTF-8-BOM und Trennzeichen `,` / `;` / Tab automatisch erkannt

### Changed

- Berechtigungen: mehrere Fallback-Gruppen (`FindGroupByNames`); Export/Invite
  nutzen `EditPersonsGroupNames` / `ExportPersonsGroupNames`
- `export -i`: zeigt immer Standort-Menü inkl. „Alle Standorte“ (nicht nur bei
  fehlendem Benutzer-Standort)

## [2.3.2.32] - 2026-06-17 10:28:25

### Fixed

- `embed-windows-icon.sh`: `go-winres` für den CI-Host installieren (nicht mit
  `GOOS=windows` — sonst `go-winres.exe`, auf Linux nicht ausführbar)

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
