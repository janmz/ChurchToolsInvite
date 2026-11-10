package cmd

import (
	"fmt"
	"os"
	"strings"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	config "github.com/janmz/churchtools-invite/internal/config"
	"github.com/janmz/churchtools-invite/internal/textnorm"
)

func interactiveExportOptions(client *churchtools.Client, cfg *config.Config) (churchtools.PersonListOptions, exportInviteFilter, error) {
	opts := churchtools.PersonListOptions{}

	choice, err := promptExportCampus(client)
	if err != nil {
		return churchtools.PersonListOptions{}, exportInviteFilterNEU, err
	}
	applyCampusChoice(&opts, choice)

	if choice.all {
		fmt.Fprintln(os.Stderr, "\nExport: alle Standorte")
	} else {
		fmt.Fprintf(os.Stderr, "\nStandort: %s\n", formatCampus(campusDisplayName(client, choice.campusID), choice.campusID))
	}

	mode, err := promptFilterMode()
	if err != nil {
		return churchtools.PersonListOptions{}, exportInviteFilterNEU, err
	}

	switch mode {
	case "status":
		statuses, err := client.ListPersonStatuses()
		if err != nil {
			return churchtools.PersonListOptions{}, exportInviteFilterNEU, err
		}
		statusItems := make([]menuItem, len(statuses))
		for i, status := range statuses {
			statusItems[i] = menuItem{id: status.ID, name: status.Name}
		}
		statusID, err := promptMenu("Personenstatus auswählen", statusItems)
		if err != nil {
			return churchtools.PersonListOptions{}, exportInviteFilterNEU, err
		}
		opts.StatusID = statusID
	case "group":
		groups, err := client.ListGroups(churchtools.GroupListOptions{CampusID: opts.CampusID})
		if err != nil {
			return churchtools.PersonListOptions{}, exportInviteFilterNEU, err
		}
		groupItems := make([]menuItem, len(groups))
		for i, group := range groups {
			groupItems[i] = menuItem{id: group.ID, name: group.Name}
		}
		groupID, err := promptMenu("Gruppe auswählen", groupItems)
		if err != nil {
			return churchtools.PersonListOptions{}, exportInviteFilterNEU, err
		}
		opts.GroupID = groupID
	}

	inviteFilter, err := promptExportInviteStatus()
	if err != nil {
		return churchtools.PersonListOptions{}, exportInviteFilterNEU, err
	}

	// stderr: with "export -o -" stdout carries only the CSV.
	summary := "\nExport: alle Standorte"
	if opts.CampusID > 0 {
		summary = "\nExport: Standort " + formatCampus(campusDisplayName(client, opts.CampusID), opts.CampusID)
	}
	if opts.StatusID > 0 {
		summary += fmt.Sprintf(", Status-ID %d", opts.StatusID)
	}
	if opts.GroupID > 0 {
		summary += fmt.Sprintf(", Gruppe-ID %d", opts.GroupID)
	}
	fmt.Fprintf(os.Stderr, "%s, %s\n", summary, describeInviteFilter(inviteFilter))

	return opts, inviteFilter, nil
}

// campusName returns the name of campus id, or "" when it is unknown.
func campusName(campuses []churchtools.Campus, id int) string {
	for _, campus := range campuses {
		if campus.ID == id {
			return campus.Name
		}
	}
	return ""
}

// formatCampus renders a campus for messages: "Nord" (ID 2), or ID 2 when
// the name is unknown.
func formatCampus(name string, id int) string {
	if name == "" {
		return fmt.Sprintf("ID %d", id)
	}
	return fmt.Sprintf("%q (ID %d)", textnorm.Printable(name), id)
}

func describeExportFilters(opts churchtools.PersonListOptions, inviteFilter exportInviteFilter) string {
	parts := make([]string, 0, 4)
	if opts.CampusID > 0 {
		parts = append(parts, fmt.Sprintf("Standort-ID %d", opts.CampusID))
	}
	if opts.StatusID > 0 {
		parts = append(parts, fmt.Sprintf("Status-ID %d", opts.StatusID))
	}
	if opts.GroupID > 0 {
		parts = append(parts, fmt.Sprintf("Gruppe-ID %d", opts.GroupID))
	}
	parts = append(parts, describeInviteFilter(inviteFilter))
	if len(parts) == 1 && inviteFilter == exportInviteFilterAll {
		return "alle Personen"
	}
	return strings.Join(parts, ", ")
}
