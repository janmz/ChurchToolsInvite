package csvfile_test

import (
	"bytes"
	"strings"
	"testing"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	csvfile "github.com/janmz/churchtools-invite/internal/csvfile"
)

func TestWriteToUsesCanonicalFormat(t *testing.T) {
	var buf bytes.Buffer
	entries := []csvfile.Entry{{
		PersonID:  42,
		FirstName: "Max",
		LastName:  "Muster",
		Email:     "max@example.org",
	}}

	if err := csvfile.WriteTo(&buf, entries); err != nil {
		t.Fatal(err)
	}

	text := buf.String()
	if !strings.HasPrefix(text, "\ufeff") {
		t.Fatal("expected UTF-8 BOM")
	}
	if !strings.Contains(text, "id,vorname,nachname,email") {
		t.Fatalf("unexpected header: %q", text)
	}
	if !strings.Contains(text, "42,Max,Muster,max@example.org") {
		t.Fatalf("unexpected row: %q", text)
	}
}

func TestEntryFromPerson(t *testing.T) {
	entry := csvfile.EntryFromPerson(churchtools.Person{
		ID:        7,
		FirstName: "Erika",
		LastName:  "Beispiel",
		Email:     "erika@example.org",
	})
	if entry.PersonID != 7 || entry.Email != "erika@example.org" {
		t.Fatalf("unexpected entry: %+v", entry)
	}
}

func TestWriteReadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/export.csv"
	entries := []csvfile.Entry{{
		PersonID:  1,
		FirstName: "Anna",
		LastName:  "Test",
		Email:     "anna@example.org",
	}}

	if err := csvfile.Write(path, entries); err != nil {
		t.Fatal(err)
	}

	loaded, err := csvfile.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0].PersonID != 1 || loaded[0].Email != "anna@example.org" {
		t.Fatalf("roundtrip failed: %+v", loaded)
	}
}
