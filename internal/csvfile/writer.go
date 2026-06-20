package csvfile

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

// ExportHeader is the canonical CSV header for export files.
var ExportHeader = []string{"id", "vorname", "nachname", "email", "standort", "status"}

// EntryFromPerson maps a ChurchTools person to an export row.
func EntryFromPerson(person churchtools.Person, campusNames map[int]string) Entry {
	return Entry{
		PersonID:  person.ID,
		FirstName: person.FirstName,
		LastName:  person.LastName,
		Campus:    campusNameForPerson(person, campusNames),
		Email:     person.Email,
		Status:    person.ExportStatusLabel(),
	}
}

func campusNameForPerson(person churchtools.Person, campusNames map[int]string) string {
	if campusNames != nil {
		if name := campusNames[person.CampusID]; name != "" {
			return name
		}
	}
	if person.CampusID > 0 {
		return strconv.Itoa(person.CampusID)
	}
	return ""
}

// EntriesFromPersons converts ChurchTools persons to export rows.
func EntriesFromPersons(persons []churchtools.Person, campusNames map[int]string) []Entry {
	entries := make([]Entry, len(persons))
	for i, person := range persons {
		entries[i] = EntryFromPerson(person, campusNames)
	}
	return entries
}

// Write stores entries in the invite CSV format (UTF-8 with BOM for Excel).
func Write(path string, entries []Entry) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("csv erstellen: %w", err)
	}
	defer file.Close()

	if err := WriteTo(file, entries); err != nil {
		return err
	}
	return file.Close()
}

// WriteTo writes entries to w using the canonical export format.
func WriteTo(w io.Writer, entries []Entry) error {
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return fmt.Errorf("bom schreiben: %w", err)
	}

	writer := csv.NewWriter(w)
	if err := writer.Write(ExportHeader); err != nil {
		return fmt.Errorf("kopfzeile schreiben: %w", err)
	}

	for _, entry := range entries {
		if err := writer.Write([]string{
			strconv.Itoa(entry.PersonID),
			entry.FirstName,
			entry.LastName,
			entry.Email,
			entry.Campus,
			entry.Status,
		}); err != nil {
			return fmt.Errorf("zeile schreiben: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("csv abschließen: %w", err)
	}
	return nil
}
