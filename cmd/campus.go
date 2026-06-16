package cmd

import (
	"fmt"
	"os"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

func applyDefaultCampus(client *churchtools.Client, opts *churchtools.PersonListOptions) error {
	if opts.CampusID > 0 {
		return nil
	}

	campusID, err := client.CurrentUserCampusID()
	if err != nil {
		return err
	}
	if campusID <= 0 {
		return nil
	}

	opts.CampusID = campusID
	name := campusDisplayName(client, campusID)
	if name != "" {
		fmt.Fprintf(os.Stderr, "Standort automatisch: %s (ID %d)\n", name, campusID)
	} else {
		fmt.Fprintf(os.Stderr, "Standort automatisch: ID %d\n", campusID)
	}
	return nil
}

func campusDisplayName(client *churchtools.Client, campusID int) string {
	campuses, err := client.ListCampuses()
	if err != nil {
		return ""
	}
	return campusName(campuses, campusID)
}
