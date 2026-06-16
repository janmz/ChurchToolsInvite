package cmd

import (
	"fmt"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	config "github.com/janmz/churchtools-invite/internal/config"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Angemeldeten ChurchTools-Benutzer anzeigen",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(runWhoAmI())
	},
}

func runWhoAmI() error {
	cfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}

	client := churchtools.NewClient(cfg.BaseURL, cfg.LoginToken, cfg.Username, cfg.Password)
	if err := client.Login(); err != nil {
		return err
	}

	user, err := client.WhoAmI()
	if err != nil {
		return err
	}

	fmt.Printf("Person-ID: %d\n", user.ID)
	fmt.Printf("Name:      %s %s\n", user.FirstName, user.LastName)
	fmt.Printf("E-Mail:    %s\n", user.Email)
	if user.CampusID > 0 {
		name := campusDisplayName(client, user.CampusID)
		if name != "" {
			fmt.Printf("Standort:  %s (ID %d)\n", name, user.CampusID)
		} else {
			fmt.Printf("Standort:  ID %d\n", user.CampusID)
		}
	}
	fmt.Printf("Instanz:   %s\n", cfg.BaseURL)
	return nil
}
