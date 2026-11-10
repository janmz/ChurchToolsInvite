package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/janmz/churchtools-invite/internal/termio"
	"github.com/janmz/churchtools-invite/internal/textnorm"
)

// stdinReader is the single buffered reader for all interactive input, so
// lines buffered by one prompt are not lost for the next (piped input).
var stdinReader = bufio.NewReader(os.Stdin)

// readLine prints prompt to stderr and returns the next stdin line without
// the line break. A last line without line break is accepted. Ctrl+C ends
// the wait with errInterrupted; after that no further input is read.
func readLine(prompt string) (string, error) {
	if err := interrupted(); err != nil {
		return "", err
	}
	fmt.Fprint(os.Stderr, prompt)

	type lineResult struct {
		line string
		err  error
	}
	ch := make(chan lineResult, 1)
	go func() {
		line, err := termio.ReadLine(stdinReader, termio.MaxLineBytes)
		ch <- lineResult{line, err}
	}()
	select {
	case res := <-ch:
		return res.line, res.err
	case <-runCtx.Done():
		return "", errInterrupted
	}
}

type menuItem struct {
	id   int
	name string
}

// promptMenu shows a numbered menu on stderr (stdout may carry CSV data) and
// returns the chosen item's id.
func promptMenu(title string, items []menuItem) (int, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("keine Einträge für %s", title)
	}

	fmt.Fprintf(os.Stderr, "\n%s:\n", title)
	for i, item := range items {
		name := textnorm.Printable(item.name)
		if item.id > 0 {
			fmt.Fprintf(os.Stderr, "  [%d] %s (ID %d)\n", i+1, name, item.id)
		} else {
			// Pseudo entries such as "Alle Standorte" have no real ID.
			fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, name)
		}
	}

	line, err := readLine("Auswahl: ")
	if err != nil {
		return 0, err
	}

	choice, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return 0, fmt.Errorf("ungültige Auswahl")
	}

	if choice < 1 || choice > len(items) {
		return 0, fmt.Errorf("Auswahl außerhalb des gültigen Bereichs")
	}
	return items[choice-1].id, nil
}

func promptExportInviteStatus() (exportInviteFilter, error) {
	fmt.Fprintln(os.Stderr, "\nEinladungsstatus:")
	fmt.Fprintln(os.Stderr, "  [n] Neu (noch nicht eingeladen)")
	fmt.Fprintln(os.Stderr, "  [e] Eingeladen")
	fmt.Fprintln(os.Stderr, "  [r] Registriert")

	line, err := readLine("Auswahl [n/e/r]: ")
	if err != nil {
		return exportInviteFilterNEU, err
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "n", "neu":
		return exportInviteFilterNEU, nil
	case "e", "eingeladen":
		return exportInviteFilterEingeladen, nil
	case "r", "registriert":
		return exportInviteFilterRegistriert, nil
	default:
		return exportInviteFilterNEU, fmt.Errorf("ungültige Auswahl für den Einladungsstatus")
	}
}

func promptFilterMode() (string, error) {
	fmt.Fprintln(os.Stderr, "\nZusätzlicher Filter:")
	fmt.Fprintln(os.Stderr, "  [a] Alle Personen am Standort")
	fmt.Fprintln(os.Stderr, "  [s] Nach Personenstatus filtern")
	fmt.Fprintln(os.Stderr, "  [g] Nach Gruppe filtern")

	line, err := readLine("Auswahl [a/s/g]: ")
	if err != nil {
		return "", err
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "a", "alle":
		return "all", nil
	case "s", "status":
		return "status", nil
	case "g", "gruppe", "group":
		return "group", nil
	default:
		return "", fmt.Errorf("ungültige Filterauswahl")
	}
}
