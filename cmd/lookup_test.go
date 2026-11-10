package cmd

import (
	"fmt"
	"strings"
	"testing"
)

var testStatuses = []namedItem{
	{ID: 1, Name: "Mitglied"},
	{ID: 2, Name: "Freund"},
	{ID: 3, Name: "Gast"},
	{ID: 7, Name: "Mitglied (ausgetreten)"},
}

func TestResolveNamedItemByIDAndName(t *testing.T) {
	cases := map[string]int{
		"3":        3,
		"gast":     3,
		"Mitglied": 1, // exact name wins over "Mitglied (ausgetreten)"
		"freu":     2, // unique substring
		"ausgetre": 7,
	}
	for value, want := range cases {
		got, err := resolveNamedItem(statusKind, testStatuses, value)
		if err != nil || got != want {
			t.Errorf("%q: got %d, err %v; want %d", value, got, err, want)
		}
	}
}

func TestResolveNamedItemErrorsListCandidates(t *testing.T) {
	_, err := resolveNamedItem(statusKind, testStatuses, "99")
	if err == nil || !strings.Contains(err.Error(), "3 = Gast") || !strings.Contains(err.Error(), "nicht gefunden") {
		t.Fatalf("unknown ID: err = %v", err)
	}
	_, err = resolveNamedItem(statusKind, testStatuses, "Besucher")
	if err == nil || !strings.Contains(err.Error(), "1 = Mitglied") {
		t.Fatalf("unknown name: err = %v", err)
	}
	_, err = resolveNamedItem(statusKind, []namedItem{{ID: 1, Name: "Mitglied A"}, {ID: 2, Name: "Mitglied B"}}, "mitglied")
	if err == nil || !strings.Contains(err.Error(), "mehrdeutig") {
		t.Fatalf("ambiguous: err = %v", err)
	}
}

func TestResolveNamedItemManyCandidatesRefersToList(t *testing.T) {
	items := make([]namedItem, maxListedInError+1)
	for i := range items {
		items[i] = namedItem{ID: i + 1, Name: fmt.Sprintf("Gruppe %d", i+1)}
	}
	_, err := resolveNamedItem(groupKind, items, "999")
	if err == nil || !strings.Contains(err.Error(), "list groups") {
		t.Fatalf("err = %v", err)
	}
}
