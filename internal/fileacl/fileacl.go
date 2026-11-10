// Package fileacl restricts files with secrets or personal data (config.json,
// person CSV) to the current user. Unix permission bits are not enforced
// on Windows, so there an explicit, non-inherited DACL is set instead.
package fileacl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// warnOutput receives warnings of RestrictOrWarn (replaceable in tests).
var warnOutput io.Writer = os.Stderr

// RestrictToOwner limits read and write access of path to the current user.
func RestrictToOwner(path string) error {
	return restrictToOwner(path)
}

// RestrictOrWarn restricts path like RestrictToOwner but only prints a
// warning when that fails (e.g. FAT drives, network shares or a file owned
// by another user), so the file stays usable.
func RestrictOrWarn(path string) {
	if err := restrictToOwner(path); err != nil {
		fmt.Fprintf(warnOutput, "Warnung: Zugriffsrechte für %s konnten nicht eingeschränkt werden (%v). Datei enthält vertrauliche Daten – Speicherort prüfen.\n", path, err)
	}
}

// WriteAtomic writes a file via a temporary file in the same directory and
// renames it into place, so an interrupted write (full disk, crash) never
// leaves a truncated file behind. A symlink at path is resolved first, so
// its target is replaced. Temporary file and result are restricted with
// RestrictOrWarn.
func WriteAtomic(path string, write func(w io.Writer) error) error {
	target := path
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		target = resolved
	}

	tmp, err := os.CreateTemp(filepath.Dir(target), "."+filepath.Base(target)+".tmp-*")
	if err != nil {
		return fmt.Errorf("temporäre Datei anlegen: %w", err)
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	// Restrict the temporary file before any data is written: on Windows the
	// 0600 mode of os.CreateTemp has no effect and the file would inherit the
	// directory's (often broad) ACL while it is being written.
	RestrictOrWarn(tmpName)
	if err := write(tmp); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("datei schreiben: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("datei schließen: %w", err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("datei ersetzen: %w", err)
	}
	committed = true
	RestrictOrWarn(target)
	return nil
}
