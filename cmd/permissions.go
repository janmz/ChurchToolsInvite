package cmd

import (
	"fmt"
	"os"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	config "github.com/janmz/churchtools-invite/internal/config"
)

func printPermissionNotes(notes []string) {
	for _, note := range notes {
		fmt.Fprintf(os.Stderr, "Berechtigung: %s\n", note)
	}
}

func ensureExportPermissions(client *churchtools.Client, cfg config.Config) error {
	notes, err := client.EnsurePermissions([]churchtools.PermissionRequirement{
		{
			Module:      churchtools.ModuleChurchDB,
			Permission:  churchtools.PermExportData,
			GroupName:   cfg.ExportPersonsGroupName(),
			Description: "Personen exportieren",
		},
	})
	if err != nil {
		return err
	}
	printPermissionNotes(notes)
	return nil
}

func ensureInvitePermissions(client *churchtools.Client, cfg config.Config, syncEmail bool) error {
	if !syncEmail {
		return nil
	}

	notes, err := client.EnsurePermissions([]churchtools.PermissionRequirement{
		{
			Module:      churchtools.ModuleChurchDB,
			Permission:  churchtools.PermWriteAccess,
			GroupName:   cfg.EditPersonsGroupName(),
			Description: "Personen bearbeiten",
		},
	})
	if err != nil {
		return err
	}
	printPermissionNotes(notes)
	return nil
}
