package csvfile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Entry represents one person row from the CSV file.
type Entry struct {
	Line      int
	PersonID  int
	FirstName string
	LastName  string
	Campus    string // export only; ignored on invite import
	Email     string
	Status    string // export only; ignored on invite import
}

var idColumns = []string{"id", "person_id", "personid", "ct_id", "churchtools_id"}
var firstNameColumns = []string{"firstname", "first_name", "vorname"}
var lastNameColumns = []string{"lastname", "last_name", "nachname"}
var emailColumns = []string{"email", "e-mail", "mail"}

// Read parses a CSV file and returns person entries.
func Read(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("csv öffnen: %w", err)
	}

	reader, err := newCSVReader(data)
	if err != nil {
		return nil, err
	}

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("csv-kopfzeile lesen: %w", err)
	}

	index := mapColumns(normalizeHeader(header))
	if _, ok := index["id"]; !ok {
		return nil, errors.New("csv benötigt eine id-spalte (id, person_id, ct_id, …)")
	}

	var entries []Entry
	line := 1
	for {
		line++
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("zeile %d lesen: %w", line, err)
		}

		if isEmptyRecord(record) {
			continue
		}

		entry, err := parseRecord(line, record, index)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return nil, errors.New("csv enthält keine datensätze")
	}
	return entries, nil
}

func mapColumns(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, name := range header {
		index[name] = i
	}

	result := make(map[string]int)
	if col, ok := findColumn(index, idColumns); ok {
		result["id"] = col
	}
	if col, ok := findColumn(index, firstNameColumns); ok {
		result["firstname"] = col
	}
	if col, ok := findColumn(index, lastNameColumns); ok {
		result["lastname"] = col
	}
	if col, ok := findColumn(index, emailColumns); ok {
		result["email"] = col
	}
	return result
}

func findColumn(index map[string]int, candidates []string) (int, bool) {
	for _, candidate := range candidates {
		if col, ok := index[candidate]; ok {
			return col, true
		}
	}
	return 0, false
}

func normalizeHeader(header []string) []string {
	out := make([]string, len(header))
	for i, field := range header {
		out[i] = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(field, "\ufeff")))
	}
	return out
}

func parseRecord(line int, record []string, index map[string]int) (Entry, error) {
	idText := strings.TrimSpace(fieldAt(record, index["id"]))
	if idText == "" {
		return Entry{}, fmt.Errorf("zeile %d: id fehlt", line)
	}

	personID, err := strconv.Atoi(idText)
	if err != nil || personID <= 0 {
		return Entry{}, fmt.Errorf("zeile %d: ungültige id %q", line, idText)
	}

	entry := Entry{
		Line:     line,
		PersonID: personID,
	}
	if col, ok := index["firstname"]; ok {
		entry.FirstName = strings.TrimSpace(fieldAt(record, col))
	}
	if col, ok := index["lastname"]; ok {
		entry.LastName = strings.TrimSpace(fieldAt(record, col))
	}
	if col, ok := index["email"]; ok {
		entry.Email = strings.TrimSpace(fieldAt(record, col))
	}
	return entry, nil
}

func fieldAt(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return record[index]
}

func isEmptyRecord(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}
