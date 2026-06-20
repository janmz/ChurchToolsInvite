package csvfile_test

import (
	"os"
	"path/filepath"
	"testing"

	csvfile "github.com/janmz/churchtools-invite/internal/csvfile"
)

func TestReadAcceptsUTF8BOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "personen.csv")
	content := "\ufeffid,vorname,nachname,email\n1001,Anna,Beispiel,anna@example.org\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	entries, err := csvfile.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].PersonID != 1001 {
		t.Fatalf("entries = %+v", entries)
	}
}

func TestReadAcceptsSemicolonDelimiter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "personen.csv")
	content := "id;vorname;nachname;email\n1001;Anna;Beispiel;anna@example.org\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	entries, err := csvfile.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Email != "anna@example.org" {
		t.Fatalf("entries = %+v", entries)
	}
}
