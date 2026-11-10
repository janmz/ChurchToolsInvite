package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	config "github.com/janmz/churchtools-invite/internal/config"
)

// ensureCampusID returns the campus of the logged-in user, the configured
// campus_id or – after asking – a chosen campus that is stored as campus_id.
func ensureCampusID(client *churchtools.Client, cfg *config.Config) (int, error) {
	campusID, err := client.CurrentUserCampusID()
	if err != nil {
		return 0, err
	}
	if campusID > 0 {
		return campusID, nil
	}
	if cfg.CampusID > 0 {
		return cfg.CampusID, nil
	}

	campuses, err := client.ListCampuses()
	if err != nil {
		return 0, err
	}
	if len(campuses) == 0 {
		// Instance without campuses: no filter is needed.
		return 0, nil
	}

	campusItems := make([]menuItem, len(campuses))
	for i, campus := range campuses {
		campusItems[i] = menuItem{id: campus.ID, name: campus.Name}
	}

	selectedID, err := promptMenu(
		"Kein Standort am Benutzer hinterlegt – bitte Standort auswählen (wird in config gespeichert)",
		campusItems,
	)
	if errors.Is(err, io.EOF) {
		return 0, fmt.Errorf("kein Standort am Benutzer hinterlegt und keine Eingabe möglich – Standort mit --campus ID|Name|all angeben")
	}
	if err != nil {
		return 0, err
	}

	cfg.CampusID = selectedID
	path := configPath
	if path == "" {
		path = config.DefaultConfigName
	}
	// Only campus_id is written; credentials from the environment stay out
	// of the file.
	if err := config.SaveCampusID(path, selectedID); err != nil {
		return 0, err
	}

	fmt.Fprintf(os.Stderr, "Standort %s in %s gespeichert.\n", formatCampus(campusName(campuses, selectedID), selectedID), path)
	return selectedID, nil
}

func campusDisplayName(client *churchtools.Client, campusID int) string {
	campuses, err := client.ListCampuses()
	if err != nil {
		return ""
	}
	return campusName(campuses, campusID)
}

func campusNameMap(client *churchtools.Client) (map[int]string, error) {
	campuses, err := client.ListCampuses()
	if err != nil {
		return nil, err
	}
	names := make(map[int]string, len(campuses))
	for _, campus := range campuses {
		names[campus.ID] = campus.Name
	}
	return names, nil
}
