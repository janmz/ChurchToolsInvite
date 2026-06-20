# ChurchTools_Invite

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![Release](https://img.shields.io/badge/Release-GitHub-0077B6)](https://github.com/janmz/ChurchToolsInvite/releases)
[![Lizenz: MIT (modifiziert)](https://img.shields.io/badge/Lizenz-MIT--Modified-blue.svg)](LICENSE)
[![Unterstützung: CFI-Kinderhilfe](https://img.shields.io/badge/Unterstützung-CFI--Kinderhilfe-0077B6?logo=heart)](https://cfi-kinderhilfe.de/jetzt-spenden?q=VAYACTINVITE)
[![Build Status](https://github.com/janmz/ChurchToolsInvite/actions/workflows/ci.yml/badge.svg)](https://github.com/janmz/ChurchToolsInvite/actions/workflows/ci.yml)

<p align="center">
  <img src="https://img.shields.io/badge/🇩🇪-Deutsch-0077B6?style=for-the-badge" alt="Deutsch (aktuell)">
  <a href="README.md"><img src="https://img.shields.io/badge/🇺🇸-English-555?style=for-the-badge" alt="English"></a>
</p>

**Churchtools-invite** ist ein schlankes Go-CLI für **Masseneinladungen zu
ChurchTools** aus einer CSV-Datei, u. a.:

- CSV-Import mit Personen-IDs und optionalem E-Mail-Abgleich
- Einladungen über die REST-API (`POST /persons/{id}/invite`)
- Personenexport mit Standort-, Einladungs- und Gruppenfiltern; standardmäßig
  nur noch nicht eingeladene Personen
- Setup, Dry-Run und Berechtigungshinweise
- Bereits eingeladene Personen standardmäßig überspringen; bei abweichender
  E-Mail in der CSV werden Adresse aktualisiert und erneut eingeladen
  (`--reinvite` nur für Status **Eingeladen**; **Registriert** wird immer
  übersprungen)

Dies soll den Prozess unterstützen, den jede Gemeinde durchmacht, wenn sie neu mit Churchtools beginnt. Meist werden Personendaten aus einem bestehenden Tool übernommen, aber alle müssen neu für Churchtools eingeladen werden, damit auch die datenschutzrechtlichlichen Einwilligungen für Chruchtools erfasst und dokumentiert werden können. Meist sind in den Altsystemen noch "Altlasten", wie verstorbene oder ausgeschiedene ehemalige Mitglieder. Oder es sind E-Mail-Adressen hinterlegt, die nicht mehr aktuell oder präferiert sind. Der letzte Umstand kann dazu führen, dass Personen doppelt angelegt werden und dann mühsam zusammengeführt werden müssen. Mit diesem Tool kann dies zu großen Teilen vermieden werden. Zuerst werden die Daten exportiert und auch eine nicht technick-affine Person kann dann die Liste bereinigen (löschen und E-Mails korrigieren). Diese Liste wird dann für die Generierungen der Einladungen verwendet, wobei bei Bedarf die E-Mail-Adressen vorher korrigiert werden. Dazu muss die Person, die das Tool bedient die entsprechenden Rechte haben, oder durch einfache Anfragen in die dazu notwendigen Gruppen kommen. Das Tool versucht bei fehlenden Rechten eine automatische Aufnahme in die notwendigen Gruppen, aber wenn dies scheitert müssen die Rechte manuell zugeordnet werden. Es ist auch möglich einen Testlauf zu unternehmen, der zeigt, wer eingeladen würde und welche Änderungen vorgenommen würden.

## Funktionen

- CSV einlesen mit Personen-IDs (`id`, `person_id`, `ct_id`, …)
- Einladungs-E-Mails über die ChurchTools-REST-API senden
- Personen exportieren (Standort, Einladungsstatus, Gruppe; interaktive Auswahl)
- Setup-Befehle für Instanzname, Login-Token, Verbindungstest und
  Berechtigungshinweise
- Dry-Run zur Prüfung von CSV und Personendaten vor dem Versand
- E-Mail aus CSV übernehmen: abweichende Adresse wird primär gesetzt,
  bisherige ChurchTools-Adresse bleibt als zusätzliche erhalten
- Bereits eingeladene Personen überspringen, sofern die E-Mail stimmt; weicht
  die CSV-Adresse ab, E-Mail aktualisieren und erneut einladen; mit
  `--reinvite` nur **Eingeladene** erneut einladen (**Registrierte** nie)
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
# config.json anpassen oder setup init

.\churchtools-invite.exe setup test
.\churchtools-invite.exe export -o personen.csv
```

**Liste manuell korrigieren!**

```bash
.\churchtools-invite.exe invite -f personen.csv --dry-run
```

**Fehler in der Liste anpassen, evtl. Rechte „besorgen“**

```bash
.\churchtools-invite.exe invite -f personen.csv
```

Globale Option: `-c config.json` für einen anderen Konfigurationspfad.

## Konfiguration

Kopiere `config.example.json` nach `config.json` oder nutze Umgebungsvariablen:

| Variable | Beschreibung |
| --- | --- |
| `CT_BASE_URL` | Instanzname (z. B. `emk-rheinmain`) oder volle URL |
| `CT_LOGIN_TOKEN` | API-Login-Token |
| `CT_USERNAME` / `CT_PASSWORD` | Alternative zum Token |
| `delay_ms` | Pause zwischen Einladungen in Millisekunden (Standard: 500) |
| `campus_id` | Standard-Standort, wenn der Benutzer keinen hat (wird bei Bedarf interaktiv gesetzt) |
| `pre_join_groups` | Kommagetrennte Gruppen, denen vor Export/Invite in dieser Reihenfolge beigetreten wird (Standard siehe `config.example.json`; `-` zum Deaktivieren) |
| `CT_PRE_JOIN_GROUPS` | Umgebungsvariable für `pre_join_groups` |

Login-Token beschaffen:

```bash
.\churchtools-invite.exe setup init
# oder nach Login:
.\churchtools-invite.exe setup token
```

### Haupt- und Nebeninstanz (OAuth)

Bei ChurchTools-Mandanten mit mehreren Standorten kann die URL einer
Nebeninstanz so aussehen: `https://haupt-neben.church.tools` (Beispiel:
`https://emk-rheinmain.church.tools`). Benutzerkonten liegen auf der
**Zentralinstanz** `https://haupt.church.tools` (Beispiel:
`https://emk.church.tools`).

Schlägt die direkte Anmeldung auf der Nebeninstanz fehl, baut das Tool bei
**Benutzername/Passwort** den OAuth-Flow nach (klappt der Direktlogin, entfällt
dieser Schritt):

1. Login auf der Zentralinstanz (`/api/login`)
2. `oauthclients/…/startlogin` auf der Nebeninstanz (Redirect zur Zentralinstanz)
3. OAuth-Authorize mit bestehender Zentral-Session
4. Callback auf der Nebeninstanz → lokale Session
5. API-Aufrufe weiter über die **Nebeninstanz** (konfigurierte URL)

`setup init` kann danach automatisch ein Login-Token der Nebeninstanz holen
(`/api/person/me/apitoken`).

Bei **Login-Token**, der nur auf der Zentralinstanz gültig ist, wird als Fallback
weiterhin die Zentralinstanz für API-Aufrufe verwendet (Hinweis in der Ausgabe).

Berechtigungen prüfen:

```bash
.\churchtools-invite.exe setup permissions
```

## CSV-Format

### Export (`export`)

Standardmäßig exportiert das Tool **nur Personen, die noch nicht eingeladen
wurden**. Mit `-i` bzw. `--invited` erscheinen auch bereits eingeladene oder
registrierte Personen in der Liste.

```csv
id,vorname,nachname,email,standort,status
123,Max,Muster,max@example.org,Rhein-Main,NEU
456,Anna,Beispiel,anna@example.org,Rhein-Main,Eingeladen
789,Tim,Test,tim@example.org,Rhein-Main,Registriert
```

- Spalte `status`: `NEU` (noch nicht eingeladen), `Eingeladen` (Einladung
  ausstehend) oder `Registriert` (ChurchTools-Konto bereits angenommen)
- Spalte `standort`: Standortname aus ChurchTools (nur Export)
- Spalten `id`, `vorname`, `nachname`, `email` sind die Einladungs-CSV;
  `standort` und `status` dienen der manuellen Bereinigung und werden beim
  Import mit `invite` ignoriert

**Standort:** `--campus` ist **nur nötig**, wenn Sie **nicht** für Ihren
eigenen Standort exportieren möchten oder Ihrem Benutzer **kein** Standort
zugeordnet ist. Ohne `--campus` nutzt `export` automatisch den Standort des
angemeldeten Nutzers bzw. `campus_id` aus `config.json`.

`--campus` akzeptiert entweder eine numerische Standort-ID oder einen
eindeutigen Namens-Teilstring. Bei Namenssuche werden Standortliste und
Suchstring in Kleinbuchstaben umgewandelt; aus dem Suchstring werden alle
Zeichen außer `a`–`z` entfernt. Gefunden wird der Standort, dessen Name den
bereinigten Suchstring enthält – es muss **genau ein** Treffer sein
(`contains(standort.name, suchstring)`).

Beispiele:

```bash
# Eigener Standort – kein --campus nötig
.\churchtools-invite.exe export -o personen.csv

# Anderer Standort per ID oder eindeutigem Namensfragment
.\churchtools-invite.exe export -o alle.csv --campus 3
.\churchtools-invite.exe export -o rheinmain.csv --campus rhein

# Auch bereits Eingeladene/Registrierte anzeigen
.\churchtools-invite.exe export -o review.csv --invited
```

### Import (`invite`)

```csv
id,vorname,nachname,email
123,Max,Muster,max@example.org
```

- Spalte `id` ist Pflicht (auch: `person_id`, `ct_id`)
- Name und E-Mail sind optional; bei abweichender E-Mail wird ChurchTools vor
  dem Einladen aktualisiert (alte Adresse bleibt als zusätzliche erhalten)
- Zusätzliche Spalten wie `standort` und `status` aus dem Export werden
  ignoriert

### Dry-run – Vor dem Versand prüfen

`invite --dry-run` durchläuft dieselbe Prüflogik wie ein echter Lauf, **ohne**
Einladungen zu senden und **ohne** Daten in ChurchTools zu ändern (kein
E-Mail-Sync). Pro CSV-Zeile wird geprüft:

- Existiert die Person-ID in ChurchTools?
- Ist die Person bereits eingeladen? Erkennung u. a. über `invitationStatus`
  (`accepted`, `pending`). Standard: überspringen, sofern die E-Mail aus der
  CSV mit ChurchTools übereinstimmt; bei abweichender E-Mail werden Update und
  erneute Einladung simuliert. **Registrierte** werden immer übersprungen;
  `--reinvite` gilt nur für **Eingeladene**
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
| `setup init` | Interaktive `config.json` anlegen (Instanzname, Passwort mit `*`-Eingabe) |
| `setup test` | Login und Verbindung testen |
| `setup token` | Login-Token anzeigen |
| `setup permissions` | Einladungs-Berechtigungen prüfen |
| `whoami` | Angemeldeten Benutzer, Standort, Gruppenmitgliedschaften (sortiert nach Name, Format `Name  ID`) und Instanz-URL anzeigen |
| `export -o DATEI` | Personenliste als Einladungs-CSV exportieren, Default `personen.csv` (`-` = stdout); standardmäßig nur noch nicht Eingeladene |
| `export -i` / `--invited` | Alle Einladungsstatus exportieren (NEU, Eingeladen, Registriert) |
| `export --interactive` | Standort, optional Filter (Personenstatus, Gruppe) und Einladungsstatus wählen (`[n]` Neu, `[e]` Eingeladen, `[r]` Registriert) |
| `export --campus WERT` | Standort per ID oder eindeutigem Namens-Teilstring (siehe CSV-Format); nur nötig bei abweichendem Standort oder fehlendem Nutzer-Standort |
| `export --campus all` | Alle Standorte (Alias zu `--all-campuses`) |
| `export --all-campuses` | Keinen Standort-Filter |
| `export --status-id ID` | Nur Personen mit diesem ChurchTools-Personenstatus (Mitglied, Gast, …) |
| `export --group-id ID` | Nur Gruppenmitglieder |
| `export --skip-permission-request` | Keine Gruppenanfrage bei fehlenden Export-Rechten |
| `export --skip-pre-join-groups` | Keine Vorab-Gruppen vor dem Export beitreten |
| `invite -f DATEI` | Einladungen senden |
| `invite -f DATEI --dry-run` | Prüfen/simulieren ohne Versand (siehe oben) |
| `invite -f DATEI --delay-ms MS` | Pause zwischen Einladungen (0 = `delay_ms` aus config) |
| `invite -f DATEI --no-sync-email` | E-Mail-Sync aus CSV deaktivieren (abweichende E-Mail → Fehler) |
| `invite -f DATEI --reinvite` | Personen mit Status **Eingeladen** erneut einladen; **Registriert** wird immer übersprungen |
| `invite -f DATEI --skip-permission-request` | Keine Gruppenanfrage bei fehlenden Rechten |
| `invite -f DATEI --skip-pre-join-groups` | Keine Vorab-Gruppen vor dem Invite beitreten |

## Entwicklung

Unter Linux/macOS bauen Release-Artefakte ohne eingebettetes Datei-Icon (kein
Standard für reine CLI-Binaries). Unter Windows wird bei Releases
`vaya.ico` im Projektroot per
[go-winres](https://github.com/tc-hib/go-winres) in die `.exe` eingebettet.

**Tests:** Im Projektroot liefert nur `go test` keine Ergebnisse (dort liegt
`package main` ohne Tests). Alle automatisierten Tests starten mit
`go test ./...` — Details, Abdeckung und manuelle Abnahme gegen eine echte
Instanz: [TESTING.de.md](TESTING.de.md).

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
den ursprünglichen Autor **Jan Neuhaus, VAYA Consulting** nennst und einen Link
auf das Original-Repository beibehältst:
`https://github.com/janmz/ChurchToolsInvite`.

**Es wird keine Gewährleistung übernommen.**

## Unterstützung

Wenn dir das Projekt nützt, unterstütze bitte die **CFI-Kinderhilfe**:
[Spendenseite](https://cfi-kinderhilfe.de/jetzt-spenden?q=VAYACTINVITE)
(Spenden gehen an die CFI-Kinderhilfe, nicht an den Autor.)

## Kontakt

**Autor**: Jan Neuhaus, VAYA Consulting –
[VAYA Consulting](https://vaya-consulting.de/development?q=GITHUB)
**Repository**: [https://github.com/janmz/ChurchToolsInvite](https://github.com/janmz/ChurchToolsInvite)

## Changelog

Siehe [Changelog.md](Changelog.md) für die Versionshistorie.
