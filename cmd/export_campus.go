package cmd

import (
	"strings"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	config "github.com/janmz/churchtools-invite/internal/config"
	"github.com/janmz/churchtools-invite/internal/textnorm"
)

const campusMenuAll = -1

type exportCampusChoice struct {
	all      bool
	campusID int
}

func parseExportCampusFlag(client *churchtools.Client, value string) (exportCampusChoice, error) {
	value = strings.TrimSpace(value)
	switch strings.ToLower(value) {
	case "all", "alle", "*":
		return exportCampusChoice{all: true}, nil
	}

	campuses, err := client.ListCampuses()
	if err != nil {
		return exportCampusChoice{}, err
	}
	id, err := matchCampusBySearch(campuses, value)
	if err != nil {
		return exportCampusChoice{}, err
	}
	return exportCampusChoice{campusID: id}, nil
}

// normalizeCampusSearch folds a campus name or search term: umlauts and
// accents are transliterated (ü → ue, é → e), letters and digits are kept,
// spaces, hyphens and punctuation are ignored.
func normalizeCampusSearch(value string) string {
	return textnorm.FoldAlnum(value)
}

// matchCampusBySearch resolves a campus ID or name (exact name first, then a
// unique substring); errors list the valid campuses.
func matchCampusBySearch(campuses []churchtools.Campus, search string) (int, error) {
	return resolveNamedItem(campusKind, campusesToItems(campuses), search)
}

func promptExportCampus(client *churchtools.Client) (exportCampusChoice, error) {
	campuses, err := client.ListCampuses()
	if err != nil {
		return exportCampusChoice{}, err
	}

	items := make([]menuItem, 0, len(campuses)+1)
	items = append(items, menuItem{id: campusMenuAll, name: "Alle Standorte"})
	for _, campus := range campuses {
		items = append(items, menuItem{id: campus.ID, name: campus.Name})
	}

	selectedID, err := promptMenu("Standort für Export", items)
	if err != nil {
		return exportCampusChoice{}, err
	}
	if selectedID == campusMenuAll {
		return exportCampusChoice{all: true}, nil
	}
	return exportCampusChoice{campusID: selectedID}, nil
}

func applyCampusChoice(opts *churchtools.PersonListOptions, choice exportCampusChoice) {
	if choice.all {
		opts.CampusID = 0
		return
	}
	opts.CampusID = choice.campusID
}

func resolveNonInteractiveCampus(
	client *churchtools.Client,
	cfg *config.Config,
	campusFlag string,
	allCampuses bool,
) (exportCampusChoice, error) {
	if allCampuses {
		return exportCampusChoice{all: true}, nil
	}
	campusFlag = strings.TrimSpace(campusFlag)
	if campusFlag != "" {
		return parseExportCampusFlag(client, campusFlag)
	}

	campusID, err := ensureCampusID(client, cfg)
	if err != nil {
		return exportCampusChoice{}, err
	}
	if campusID <= 0 {
		return exportCampusChoice{all: true}, nil
	}
	return exportCampusChoice{campusID: campusID}, nil
}
