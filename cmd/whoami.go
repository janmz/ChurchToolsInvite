package cmd

import (
	"fmt"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	config "github.com/janmz/churchtools-invite/internal/config"
	"github.com/janmz/churchtools-invite/internal/textnorm"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Angemeldeten ChurchTools-Benutzer anzeigen",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWhoAmI()
	},
}

func runWhoAmI() error {
	cfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}

	client, err := connectChurchTools(cfg)
	if err != nil {
		return err
	}

	user, err := client.WhoAmI()
	if err != nil {
		return err
	}

	fmt.Printf("Person-ID:   %d\n", user.ID)
	fmt.Printf("Name:        %s %s\n", textnorm.Printable(user.FirstName), textnorm.Printable(user.LastName))
	fmt.Printf("E-Mail:      %s\n", textnorm.Printable(user.Email))

	if user.CampusID > 0 {
		fmt.Printf("Standort-ID: %d\n", user.CampusID)
		if name := campusDisplayName(client, user.CampusID); name != "" {
			fmt.Printf("Standort:    %s\n", textnorm.Printable(name))
		}
	} else {
		fmt.Println("Standort-ID: nicht zugeordnet")
		if cfg.CampusID > 0 {
			fmt.Printf("Standard-Standort (config): %s\n", formatCampus(campusDisplayName(client, cfg.CampusID), cfg.CampusID))
		}
	}

	groups, err := client.ListPersonGroups(user.ID)
	if err != nil {
		return err
	}
	churchtools.SortGroupsByName(groups)
	if len(groups) == 0 {
		fmt.Println("Gruppen:     (keine)")
	} else {
		fmt.Println("Gruppen:")
		for _, group := range groups {
			fmt.Printf("  %s  %d\n", textnorm.Printable(churchtools.PlainGroupName(group.Name)), group.ID)
		}
	}

	fmt.Printf("Instanz:     %s\n", client.BaseURL())
	return nil
}
