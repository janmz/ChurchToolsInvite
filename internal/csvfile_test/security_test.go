package csvfile_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	csvfile "github.com/janmz/churchtools-invite/internal/csvfile"
)

func TestWriteEscapesFormulaCellsAndReadRestoresThem(t *testing.T) {
	entries := []csvfile.Entry{{
		PersonID:  7,
		FirstName: "=HYPERLINK(\"http://example.org\")",
		LastName:  "+Muster",
		Email:     "-max@example.org",
		Campus:    "-Nord",
		Status:    "NEU",
	}}

	var buf bytes.Buffer
	if err := csvfile.WriteTo(&buf, entries); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{`"'=HYPERLINK(""http://example.org"")"`, "'+Muster", "'-max@example.org", "'-Nord"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output lacks escaped cell %q:\n%s", want, out)
		}
	}

	path := filepath.Join(t.TempDir(), "personen.csv")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	read, err := csvfile.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if read[0].FirstName != entries[0].FirstName || read[0].LastName != "+Muster" || read[0].Email != "-max@example.org" {
		t.Fatalf("round trip = %+v", read[0])
	}
}

func TestReadKeepsLegitimateApostrophe(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personen.csv")
	content := "id,vorname,nachname,email\n1,'Anna,O'Brien,anna@example.org\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	entries, err := csvfile.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].FirstName != "'Anna" || entries[0].LastName != "O'Brien" {
		t.Fatalf("entry = %+v", entries[0])
	}
}

func TestReadRejectsDuplicatePersonID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personen.csv")
	content := "id,vorname,nachname,email\n5,Anna,A,a@example.org\n6,Ben,B,b@example.org\n5,Anna,A,a2@example.org\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := csvfile.Read(path)
	if err == nil || !strings.Contains(err.Error(), "zeile 4") || !strings.Contains(err.Error(), "Zeile 2") {
		t.Fatalf("err = %v, want duplicate error naming lines 4 and 2", err)
	}
}

// S5: invalid e-mail values are rejected with the line number.
func TestReadRejectsInvalidEmail(t *testing.T) {
	for _, email := range []string{"a@x.de; b@y.de", "Max <max@example.org>", "kein-at", "max @example.org"} {
		path := filepath.Join(t.TempDir(), "personen.csv")
		content := "id,vorname,nachname,email\n5,Anna,A,\"" + email + "\"\n"
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := csvfile.Read(path); err == nil || !strings.Contains(err.Error(), "zeile 2") {
			t.Fatalf("email %q: err = %v", email, err)
		}
	}
}
