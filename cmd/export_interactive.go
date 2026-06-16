package cmd

import (
	"fmt"
	"strings"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

func interactiveExportOptions(client *churchtools.Client) (churchtools.PersonListOptions, error) {
	opts := churchtools.PersonListOptions{}

	campusID, err := client.CurrentUserCampusID()
	if err != nil {
		return churchtools.PersonListOptions{}, err
	}

	if campusID > 0 {
		opts.CampusID = campusID
		name := campusDisplayName(client, campusID)
		if name != "" {
			fmt.Printf("\nStandort automatisch: %s (ID %d)\n", name, campusID)
		} else {
			fmt.Printf("\nStandort automatisch: ID %d\n", campusID)
		}
	} else {
		campuses, err := client.ListCampuses()
		if err != nil {
			return churchtools.PersonListOptions{}, err
		}

		campusItems := make([]menuItem, len(campuses))
		for i, campus := range campuses {
			campusItems[i] = menuItem{id: campus.ID, name: campus.Name}
		}

		campusID, err = promptMenu("Standort auswählen", campusItems, false)
		if err != nil {
			return churchtools.PersonListOptions{}, err
		}
		opts.CampusID = campusID
	}

	mode, err := promptFilterMode()
	if err != nil {
		return churchtools.PersonListOptions{}, err
	}

	switch mode {
	case "status":
		statuses, err := client.ListPersonStatuses()
		if err != nil {
			return churchtools.PersonListOptions{}, err
		}
		statusItems := make([]menuItem, len(statuses))
		for i, status := range statuses {
			statusItems[i] = menuItem{id: status.ID, name: status.Name}
		}
		statusID, err := promptMenu("Personenstatus auswählen", statusItems, false)
		if err != nil {
			return churchtools.PersonListOptions{}, err
		}
		opts.StatusID = statusID
	case "group":
		groups, err := client.ListGroups(churchtools.GroupListOptions{CampusID: campusID})
		if err != nil {
			return churchtools.PersonListOptions{}, err
		}
		groupItems := make([]menuItem, len(groups))
		for i, group := range groups {
			groupItems[i] = menuItem{id: group.ID, name: group.Name}
		}
		groupID, err := promptMenu("Gruppe auswählen", groupItems, false)
		if err != nil {
			return churchtools.PersonListOptions{}, err
		}
		opts.GroupID = groupID
	}

	if opts.CampusID > 0 {
		selectedCampus := campusDisplayName(client, opts.CampusID)
		if selectedCampus == "" {
			selectedCampus = fmt.Sprintf("ID %d", opts.CampusID)
		}
		fmt.Printf("\nExport: Standort %q (ID %d)", selectedCampus, opts.CampusID)
	} else {
		fmt.Print("\nExport: alle Standorte")
	}
	if opts.StatusID > 0 {
		fmt.Printf(", Status-ID %d", opts.StatusID)
	}
	if opts.GroupID > 0 {
		fmt.Printf(", Gruppe-ID %d", opts.GroupID)
	}
	fmt.Println()

	return opts, nil
}

func campusName(campuses []churchtools.Campus, id int) string {
	for _, campus := range campuses {
		if campus.ID == id {
			return campus.Name
		}
	}
	return fmt.Sprintf("ID %d", id)
}

func describeExportFilters(opts churchtools.PersonListOptions) string {
	parts := make([]string, 0, 3)
	if opts.CampusID > 0 {
		parts = append(parts, fmt.Sprintf("Standort-ID %d", opts.CampusID))
	}
	if opts.StatusID > 0 {
		parts = append(parts, fmt.Sprintf("Status-ID %d", opts.StatusID))
	}
	if opts.GroupID > 0 {
		parts = append(parts, fmt.Sprintf("Gruppe-ID %d", opts.GroupID))
	}
	if len(parts) == 0 {
		return "alle Personen"
	}
	return strings.Join(parts, ", ")
}
