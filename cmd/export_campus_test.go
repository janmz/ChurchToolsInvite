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

func TestParseExportCampusFlagNumeric(t *testing.T) {
	choice, err := parseExportCampusFlag(nil, "42")
	if err != nil {
		t.Fatal(err)
	}
	if choice.all || choice.campusID != 42 {
		t.Fatalf("choice = %+v", choice)
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
