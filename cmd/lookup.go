package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/janmz/churchtools-invite/internal/textnorm"
)

// namedItem is a ChurchTools entity that can be selected by ID or name
// (campus, person status, group).
type namedItem struct {
	ID   int
	Name string
}

// lookupKind describes the entity for messages and the matching list command.
type lookupKind struct {
	flag    string // e.g. "--status"
	label   string // e.g. "Personenstatus"
	listCmd string // e.g. "list statuses"
}

// maxListedInError is the number of entries an error message lists before
// it only refers to the list command.
const maxListedInError = 25

// resolveNamedItem resolves value (an ID or a name) against items. Names are
// compared after folding (umlauts, accents, case, spaces and punctuation
// ignored): an exact name wins, otherwise a unique substring is used. Unknown
// IDs, missing or ambiguous names produce an error that names the candidates.
func resolveNamedItem(kind lookupKind, items []namedItem, value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("%s: leerer Wert", kind.flag)
	}

	if id, err := strconv.Atoi(value); err == nil {
		if id <= 0 {
			return 0, fmt.Errorf("%s muss positiv sein", kind.flag)
		}
		for _, item := range items {
			if item.ID == id {
				return id, nil
			}
		}
		return 0, fmt.Errorf("%s: %s mit ID %d nicht gefunden (oder nicht sichtbar). %s",
			kind.flag, kind.label, id, describeCandidates(kind, items))
	}

	needle := textnorm.FoldAlnum(value)
	if needle == "" {
		return 0, fmt.Errorf("%s: %q enthält keine Buchstaben oder Ziffern", kind.flag, value)
	}
	var exact, partial []namedItem
	for _, item := range items {
		name := textnorm.FoldAlnum(item.Name)
		switch {
		case name == needle:
			exact = append(exact, item)
		case strings.Contains(name, needle):
			partial = append(partial, item)
		}
	}
	matches := exact
	if len(matches) == 0 {
		matches = partial
	}

	switch len(matches) {
	case 0:
		return 0, fmt.Errorf("%s: kein %s passt zu %q. %s", kind.flag, kind.label, value, describeCandidates(kind, items))
	case 1:
		return matches[0].ID, nil
	default:
		return 0, fmt.Errorf("%s: %q ist mehrdeutig – bitte genauer angeben oder die ID verwenden: %s",
			kind.flag, value, formatItems(matches))
	}
}

// describeCandidates lists the valid entries, or refers to the list command
// when there are too many.
func describeCandidates(kind lookupKind, items []namedItem) string {
	switch {
	case len(items) == 0:
		return fmt.Sprintf("Keine Einträge sichtbar (Berechtigung?). Übersicht: %s", kind.listCmd)
	case len(items) > maxListedInError:
		return fmt.Sprintf("%d Einträge vorhanden – Übersicht mit ID und Name: %s", len(items), kind.listCmd)
	default:
		return "Gültig: " + formatItems(items)
	}
}

func formatItems(items []namedItem) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = fmt.Sprintf("%d = %s", item.ID, textnorm.Printable(item.Name))
	}
	return strings.Join(parts, ", ")
}
