//go:build windows

package fileacl

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestRestrictToOwnerLeavesSingleACE(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := RestrictToOwner(path); err != nil {
		t.Fatal(err)
	}

	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if dacl == nil || dacl.AceCount != 1 {
		t.Fatalf("DACL has %d entries, want exactly 1 (current user)", dacl.AceCount)
	}
	control, _, err := sd.Control()
	if err != nil {
		t.Fatal(err)
	}
	if control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatal("DACL must be protected against inherited entries")
	}
}

// S4: the temporary file is restricted before data is written to it.
func TestWriteAtomicRestrictsTempFileBeforeWriting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "personen.csv")
	var aces uint16
	err := WriteAtomic(path, func(w io.Writer) error {
		file, ok := w.(*os.File)
		if !ok {
			t.Fatalf("writer is %T, want *os.File", w)
		}
		sd, err := windows.GetNamedSecurityInfo(file.Name(), windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			return err
		}
		dacl, _, err := sd.DACL()
		if err != nil {
			return err
		}
		aces = dacl.AceCount
		_, err = w.Write([]byte("id\n1\n"))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if aces != 1 {
		t.Fatalf("temporary file DACL has %d entries while writing, want 1", aces)
	}
}
