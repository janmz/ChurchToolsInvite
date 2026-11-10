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

func ensurePreJoinGroups(client *churchtools.Client, cfg config.Config) error {
	names := cfg.PreJoinGroupNames()
	if len(names) == 0 {
		return nil
	}

	results, err := client.EnsurePreJoinGroups(names)
	if err != nil {
		return err
	}

	for _, result := range results {
		// The status decides; Skipped only says the membership existed
		// before (a pending or to-be-deleted membership grants no rights).
		switch result.Status {
		case churchtools.MembershipActive:
			msg := result.Message
			if msg == "" {
				msg = "Mitgliedschaft aktiv (nur für diesen Lauf)"
			}
			fmt.Fprintf(os.Stderr, "Vorab-Gruppe %q: %s\n", result.GroupName, msg)
		case churchtools.MembershipRequested:
			fmt.Fprintf(os.Stderr, "Vorab-Gruppe %q: beantragt", result.GroupName)
			if result.Message != "" {
				fmt.Fprintf(os.Stderr, " (%s)", result.Message)
			}
			fmt.Fprintln(os.Stderr)
		default:
			msg := result.Message
			if msg == "" {
				msg = "Beitritt nicht möglich"
			}
			fmt.Fprintf(os.Stderr, "Vorab-Gruppe %q: %s\n", result.GroupName, msg)
		}
	}
	return nil
}

func ensureExportPermissions(client *churchtools.Client) error {
	notes, err := client.EnsurePermissions([]churchtools.PermissionRequirement{
		{
			Module:      churchtools.ModuleChurchDB,
			Permission:  churchtools.PermExportData,
			GroupNames:  config.DefaultExportPersonsGroups,
			Description: "Personen exportieren",
		},
	})
	if err != nil {
		return err
	}
	printPermissionNotes(notes)
	return nil
}

func ensureInvitePermissions(client *churchtools.Client, syncEmail bool) error {
	if !syncEmail {
		return nil
	}

	notes, err := client.EnsurePermissions([]churchtools.PermissionRequirement{
		{
			Module:      churchtools.ModuleChurchDB,
			Permission:  churchtools.PermWriteAccess,
			GroupNames:  config.DefaultEditPersonsGroups,
			Description: "Personen bearbeiten",
		},
	})
	if err != nil {
		return err
	}
	printPermissionNotes(notes)
	return nil
}
