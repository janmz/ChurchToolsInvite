package csvfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/janmz/masseneinladung/internal/csvfile"
)

func TestReadCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "personen.csv")
	content := "id,vorname,nachname,email\n42,Max,Muster,max@example.org\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err := csvfile.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len = %d", len(entries))
	}
	if entries[0].PersonID != 42 || entries[0].FirstName != "Max" || entries[0].Email != "max@example.org" {
		t.Fatalf("unexpected entry: %+v", entries[0])
	}
}

func TestReadCSVMissingIDColumn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.csv")
	if err := os.WriteFile(path, []byte("name,email\nMax,max@example.org\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := csvfile.Read(path); err == nil {
		t.Fatal("expected error for missing id column")
	}
}
