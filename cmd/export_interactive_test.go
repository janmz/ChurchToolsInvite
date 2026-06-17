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
		name string
		opts churchtools.PersonListOptions
		want string
	}{
		{name: "all", opts: churchtools.PersonListOptions{}, want: "alle Personen"},
		{
			name: "campus only",
			opts: churchtools.PersonListOptions{CampusID: 3},
			want: "Standort-ID 3",
		},
		{
			name: "combined",
			opts: churchtools.PersonListOptions{CampusID: 1, StatusID: 5, GroupID: 9},
			want: "Standort-ID 1, Status-ID 5, Gruppe-ID 9",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := describeExportFilters(tt.opts)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
