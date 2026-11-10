package fileacl

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// CR-17: a failed write must leave the previous file untouched and no
// temporary file behind.
func TestWriteAtomicKeepsOldFileOnError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("alt"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := WriteAtomic(path, func(w io.Writer) error {
		_, _ = w.Write([]byte("halb"))
		return errors.New("abbruch")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "alt" {
		t.Fatalf("old content = %q, %v", data, err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("temporary file left behind: %v", entries)
	}
}

func TestWriteAtomicReplacesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personen.csv")
	if err := os.WriteFile(path, []byte("alt"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomic(path, func(w io.Writer) error {
		_, err := w.Write([]byte("neu"))
		return err
	}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "neu" {
		t.Fatalf("content = %q, %v", data, err)
	}
}
