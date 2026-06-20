package cmd

import (
	"testing"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

func TestCampusName(t *testing.T) {
	t.Parallel()

	campuses := []churchtools.Campus{
		{ID: 1, Name: "Mitte"},
		{ID: 2, Name: "Nord"},
	}

	if got := campusName(campuses, 2); got != "Nord" {
		t.Fatalf("got %q", got)
	}
	if got := campusName(campuses, 99); got != "ID 99" {
		t.Fatalf("got %q", got)
	}
}

func TestDescribeExportFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		opts         churchtools.PersonListOptions
		inviteFilter exportInviteFilter
		want         string
	}{
		{name: "all", opts: churchtools.PersonListOptions{}, inviteFilter: exportInviteFilterNEU, want: "nur NEU"},
		{name: "all invited", opts: churchtools.PersonListOptions{}, inviteFilter: exportInviteFilterAll, want: "alle Personen"},
		{
			name:         "campus only",
			opts:         churchtools.PersonListOptions{CampusID: 3},
			inviteFilter: exportInviteFilterNEU,
			want:         "Standort-ID 3, nur NEU",
		},
		{
			name:         "combined",
			opts:         churchtools.PersonListOptions{CampusID: 1, StatusID: 5, GroupID: 9},
			inviteFilter: exportInviteFilterNEU,
			want:         "Standort-ID 1, Status-ID 5, Gruppe-ID 9, nur NEU",
		},
		{
			name:         "eingeladen",
			opts:         churchtools.PersonListOptions{CampusID: 2},
			inviteFilter: exportInviteFilterEingeladen,
			want:         "Standort-ID 2, nur Eingeladen",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := describeExportFilters(tt.opts, tt.inviteFilter)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFilterExportPersonsByInviteStatus(t *testing.T) {
	t.Parallel()

	pending := churchtools.Person{ID: 1, InvitationStatus: "pending"}
	accepted := churchtools.Person{ID: 2, InvitationStatus: "accepted"}
	neu := churchtools.Person{ID: 3}

	persons := []churchtools.Person{pending, accepted, neu}

	if got := len(filterExportPersons(persons, exportInviteFilterNEU)); got != 1 {
		t.Fatalf("NEU filter = %d", got)
	}
	if got := len(filterExportPersons(persons, exportInviteFilterEingeladen)); got != 1 {
		t.Fatalf("Eingeladen filter = %d", got)
	}
	if got := len(filterExportPersons(persons, exportInviteFilterRegistriert)); got != 1 {
		t.Fatalf("Registriert filter = %d", got)
	}
	if got := len(filterExportPersons(persons, exportInviteFilterAll)); got != 3 {
		t.Fatalf("all filter = %d", got)
	}
}
