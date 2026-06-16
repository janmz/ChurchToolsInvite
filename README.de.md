# ChurchTools_Invite

[![Go Version](https://img.shields.io/github/go-mod/go-version/janmz/ChurchToolsInvite)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/janmz/ChurchToolsInvite)](https://github.com/janmz/ChurchToolsInvite/releases)
[![Lizenz: MIT (modifiziert)](https://img.shields.io/badge/Lizenz-MIT--Modified-blue.svg)](LICENSE)
[![Unterstützung: CFI-Kinderhilfe](https://img.shields.io/badge/Unterstützung-CFI--Kinderhilfe-0077B6?logo=heart)](https://cfi-kinderhilfe.de/jetzt-spenden?q=VAYAMASSEN)
[![Build Status](https://github.com/janmz/ChurchToolsInvite/actions/workflows/ci.yml/badge.svg)](https://github.com/janmz/ChurchToolsInvite/actions/workflows/ci.yml)

<p align="center">
  <img src="https://img.shields.io/badge/🇩🇪-Deutsch-0077B6?style=for-the-badge" alt="Deutsch (aktuell)">
  <a href="README.md"><img src="https://img.shields.io/badge/🇺🇸-English-555?style=for-the-badge" alt="English"></a>
</p>

**Churchtools-invite** ist ein schlankes Go-CLI für **Masseneinladungen zu
ChurchTools** aus einer CSV-Datei, u. a.:

- CSV-Import mit Personen-IDs und optionalem E-Mail-Abgleich
- Einladungen über die REST-API (`POST /persons/{id}/invite`)
- Personenexport mit Standort-, Status- und Gruppenfiltern
- Setup, Dry-Run und Berechtigungshinweise
- Bereits eingeladene Personen standardmäßig überspringen (`--reinvite` zum
  erneuten Einladen)

Dies soll den Prozess unterstützen, den jede Gemeinde durchmacht, wenn sie neu mit Churchtools beginnt. Meist werden Personendaten aus einem bestehenden Tool übernommen, aber alle müssen neu für Churchtools eingeladen werden, damit auch die datenschutzrechtlichlichen Einwilligungen für Chruchtools erfasst und dokumentiert werden können. Meist sind in den Altsystemen noch "Altlasten", wie verstorbene oder ausgeschiedene ehemalige Mitglieder. Oder es sind E-Mail-Adressen hinterlegt, die nicht mehr aktuell oder präferiert sind. Der letzte Umstand kann dazu führen, dass Personen doppelt angelegt werden und dann mühsam zusammengeführt werden müssen. Mit diesem Tool kann dies zu großen Teilen vermieden werden. Zuerst werden die Daten exportiert und auch eine nicht technick-affine Person kann dann die Liste bereinigen (löschen und E-Mails korrigieren). Diese Liste wird dann für die Generierungen der Einladungen verwendet, wobei bei Bedarf die E-Mail-Adressen vorher korrigiert werden. Dazu muss die Person, die das Tool bedient die entsprechenden Rechte haben, oder durch einfache Anfragen in die dazu notwendigen Gruppen kommen. Das Tool versucht bei fehlenden Rechten eine automatische Aufnahme in die notwendigen Gruppen, aber wenn dies scheitert müssen die Rechte manuell zugeordnet werden. Es ist auch möglich einen Testlauf zu unternehmen, der zeigt, wer eingeladen würde und welche Änderungen vorgenommen würden.

## Funktionen

- CSV einlesen mit Personen-IDs (`id`, `person_id`, `ct_id`, …)
- Einladungs-E-Mails über die ChurchTools-REST-API senden
- Personen exportieren (Standort, Status, Gruppe; interaktive Auswahl)
- Setup-Befehle für URL, Login-Token, Verbindungstest und Berechtigungshinweise
- Dry-Run zur Prüfung von CSV und Personendaten vor dem Versand
- E-Mail aus CSV übernehmen: abweichende Adresse wird primär gesetzt,
  bisherige ChurchTools-Adresse bleibt als zusätzliche erhalten
- Bereits eingeladene Personen standardmäßig überspringen; mit `--reinvite`
  erneut einladen
- Automatische Gruppenanfrage bei fehlenden Rechten für Export und
  E-Mail-Sync

## Voraussetzungen

- Go 1.22+ (zum Bauen aus dem Quellcode)
- ChurchTools-Konto mit Berechtigung **Personen zur Nutzung von ChurchTools
  einladen**
- Für Export: Berechtigung **export data** (Gruppe „Personen exportieren“)
- Für E-Mail-Sync beim Einladen: Berechtigung **write access** (Gruppe
  „Personen bearbeiten“)
- Login-Token oder Benutzername/Passwort

## Installation

### Binary herunterladen

Fertige Builds für Linux, macOS und Windows:
[Releases](https://github.com/janmz/ChurchToolsInvite/releases)

Archiv entpacken, `churchtools-invite` (bzw. `churchtools-invite.exe` unter
Windows) ausführen.

### Go Install

```bash
go install github.com/janmz/churchtools-invite@latest
```

### Aus Quellcode bauen

```bash
git clone https://github.com/janmz/ChurchToolsInvite.git
cd ChurchToolsInvite
go build -o churchtools-invite.exe .
```

Unter Linux/macOS heißt die Datei `churchtools-invite` (ohne `.exe`).

## Verwendung

### Schnellstart

```bash
copy config.example.json config.json
# config.json anpassen


```bash
.\churchtools-invite.exe setup test
.\churchtools-invite.exe export -o personen.csv
```

**Liste manuell korrigieren!**

```bash
.\churchtools-invite.exe invite -f personen.csv --dry-run
```

**Fehler in der Liste anpassen, evtl. Rechte "besorgen"**

```bash
.\churchtools-invite.exe invite -f personen.csv
```

Globale Option: `-c config.json` für einen anderen Konfigurationspfad.

## Konfiguration

Kopiere `config.example.json` nach `config.json` oder nutze Umgebungsvariablen:

| Variable | Beschreibung |
| --- | --- |
| `CT_BASE_URL` | URL der ChurchTools-Instanz |
| `CT_LOGIN_TOKEN` | API-Login-Token |
| `CT_USERNAME` / `CT_PASSWORD` | Alternative zum Token |
| `delay_ms` | Pause zwischen Einladungen in Millisekunden (Standard: 500) |
| `permission_groups.edit_persons` | Gruppe für Schreibzugriff (Standard: Personen bearbeiten) |
| `permission_groups.export_persons` | Gruppe für Export (Standard: Personen exportieren) |

Login-Token beschaffen:

```bash
.\churchtools-invite.exe setup init
# oder nach Login:
.\churchtools-invite.exe setup token
```

Berechtigungen prüfen:

```bash
.\churchtools-invite.exe setup permissions
```

## CSV-Format

```csv
id,vorname,nachname,email
123,Max,Muster,max@example.org
```

- Spalte `id` ist Pflicht (auch: `person_id`, `ct_id`)
- Name und E-Mail sind optional; bei abweichender E-Mail wird ChurchTools vor
  dem Einladen aktualisiert (alte Adresse bleibt als zusätzliche erhalten)

### Dry-run – Vor dem Versand prüfen

`invite --dry-run` durchläuft dieselbe Prüflogik wie ein echter Lauf, **ohne**
Einladungen zu senden und **ohne** Daten in ChurchTools zu ändern (kein
E-Mail-Sync). Pro CSV-Zeile wird geprüft:

- Existiert die Person-ID in ChurchTools?
- Ist die Person bereits eingeladen? (Standard: als „übersprungen“ gemeldet)
- Liegt eine Einladungs-E-Mail vor (CSV und/oder ChurchTools)?
- Wäre ein E-Mail-Abgleich aus der CSV erforderlich?

Ausgabe: Zeilenweises Protokoll mit `OK`, `ÜBERSPRUNGEN` oder `FEHLER` plus
Zusammenfassung. Bei mindestens einem Fehler endet das Programm mit Exit-Code
1.

Empfohlen vor dem ersten echten Versand. Alle Optionen von `invite`
(`--reinvite`, `--no-sync-email`, …) gelten auch im Dry-Run.

## Befehle

| Befehl | Zweck |
| --- | --- |
| `setup init` | Interaktive `config.json` anlegen |
| `setup test` | Login und Verbindung testen |
| `setup token` | Login-Token anzeigen |
| `setup permissions` | Einladungs-Berechtigungen prüfen |
| `whoami` | Angemeldeten Benutzer anzeigen |
| `export -o DATEI` | Personenliste als Einladungs-CSV exportieren, Default `personen.csv` |
| `export -i` | Standort und Filter interaktiv wählen |
| `export --campus-id ID` | Nur Personen dieses Standorts |
| `export --all-campuses` | Keinen Standort-Filter (Standard: Standort des Nutzers) |
| `export --status-id ID` | Nur Personen mit diesem Status |
| `export --group-id ID` | Nur Gruppenmitglieder |
| `invite -f DATEI` | Einladungen senden |
| `invite -f DATEI --dry-run` | Prüfen/simulieren ohne Versand (siehe oben) |
| `invite -f DATEI --no-sync-email` | E-Mail-Sync aus CSV deaktivieren |
| `invite -f DATEI --reinvite` | Bereits eingeladene Personen erneut einladen |
| `invite -f DATEI --skip-permission-request` | Keine Gruppenanfrage bei fehlenden Rechten |

## Entwicklung

```bash
go test ./...
go vet ./...
go build -o churchtools-invite.exe .
```

## Contributing

Beiträge sind willkommen! Bitte vor einem Pull Request
[CONTRIBUTING.de.md](CONTRIBUTING.de.md) lesen.

## Lizenz

Diese Software steht unter einer modifizierten MIT-Lizenz (siehe [LICENSE](LICENSE)).
Du darfst den Code frei verwenden, anpassen und weitergeben, **solange** du
den ursprünglichen Autor **Jan Neuhaus** nennst und einen Link auf das
Original-Repository beibehältst: `https://github.com/janmz/ChurchToolsInvite`.

**Es wird keine Gewährleistung übernommen.**

## Unterstützung

Wenn dir das Projekt nützt, unterstütze bitte die **CFI-Kinderhilfe**:
[Spendenseite](https://cfi-kinderhilfe.de/jetzt-spenden?q=VAYACTINVITE)
(Spenden gehen an die CFI-Kinderhilfe, nicht an den Autor.)

## Kontakt

**Autor**: Jan Neuhaus – [VAYA Consulting](https://vaya-consulting.de/development?q=GITHUB)
**Repository**: [https://github.com/janmz/ChurchToolsInvite](https://github.com/janmz/ChurchToolsInvite)

## Changelog

Siehe [Changelog.md](Changelog.md) für die Versionshistorie.
