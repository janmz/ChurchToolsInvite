package cmd

import (
	"testing"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

func TestParseExportCampusFlagAll(t *testing.T) {
	for _, value := range []string{"all", "ALL", "alle", "*"} {
		choice, err := parseExportCampusFlag(nil, value)
		if err != nil {
			t.Fatalf("parseExportCampusFlag(%q): %v", value, err)
		}
		if !choice.all || choice.campusID != 0 {
			t.Fatalf("choice = %+v, want all", choice)
		}
	}
}

func TestMatchCampusBySearchNumericIsValidated(t *testing.T) {
	campuses := []churchtools.Campus{{ID: 42, Name: "Nord"}}
	if id, err := matchCampusBySearch(campuses, "42"); err != nil || id != 42 {
		t.Fatalf("id = %d, err = %v", id, err)
	}
	if _, err := matchCampusBySearch(campuses, "7"); err == nil {
		t.Fatal("unknown campus ID accepted")
	}
}

func TestNormalizeCampusSearch(t *testing.T) {
	if got := normalizeCampusSearch(" Rhein-Main "); got != "rheinmain" {
		t.Fatalf("got %q", got)
	}
}

func TestMatchCampusBySearch(t *testing.T) {
	campuses := []churchtools.Campus{
		{ID: 1, Name: "EMK Mitte"},
		{ID: 2, Name: "EMK Rhein-Main"},
		{ID: 3, Name: "EMK Nord"},
	}

	id, err := matchCampusBySearch(campuses, "rhein")
	if err != nil || id != 2 {
		t.Fatalf("id = %d, err = %v", id, err)
	}

	if _, err := matchCampusBySearch(campuses, "emk"); err == nil {
		t.Fatal("expected ambiguous match")
	}

	if _, err := matchCampusBySearch(campuses, "sued"); err == nil {
		t.Fatal("expected no match")
	}
}

func TestFilterExportPersons(t *testing.T) {
	persons := []churchtools.Person{
		{ID: 1},
		{ID: 2, InvitationStatus: "pending"},
		{ID: 3, InvitationStatus: "accepted"},
	}

	filtered := filterExportPersons(persons, exportInviteFilterNEU)
	if len(filtered) != 1 || filtered[0].ID != 1 {
		t.Fatalf("filtered = %+v", filtered)
	}

	all := filterExportPersons(persons, exportInviteFilterAll)
	if len(all) != 3 {
		t.Fatalf("all = %+v", all)
	}
}

func TestMatchCampusBySearchUmlautsAndDigits(t *testing.T) {
	campuses := []churchtools.Campus{
		{ID: 1, Name: "Standort 2"},
		{ID: 2, Name: "Standort 3"},
		{ID: 3, Name: "Süd"},
		{ID: 4, Name: "Sundern"},
		{ID: 5, Name: "Café Noël"},
	}

	cases := map[string]int{
		"standort 2": 1, // digits distinguish the campuses
		"Standort3":  2,
		"Süd":        3, // ü → ue, must not match "Sundern"
		"sued":       3,
		"cafe noel":  5, // accents removed
	}
	for search, want := range cases {
		id, err := matchCampusBySearch(campuses, search)
		if err != nil || id != want {
			t.Errorf("search %q: id = %d, err = %v, want %d", search, id, err, want)
		}
	}
}

func TestMatchCampusBySearchPrefersExactName(t *testing.T) {
	campuses := []churchtools.Campus{
		{ID: 1, Name: "Nord"},
		{ID: 2, Name: "Nordost"},
	}
	id, err := matchCampusBySearch(campuses, "nord")
	if err != nil || id != 1 {
		t.Fatalf("id = %d, err = %v, want exact match 1", id, err)
	}
}
