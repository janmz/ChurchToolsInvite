package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "Churchtools-Invite",
	Short: "ChurchTools-Einladungen aus CSV versenden",
	Long: `ChurchTools-Invite liest Personen-IDs aus einer CSV-Datei und
versendet über die ChurchTools-API Einladungs-E-Mails.

Nutze 'setup' für Ersteinrichtung von URL, Login-Token und Berechtigungsprüfung.`,
	Version: "undefined",
}

// Execute runs the root command.
func Execute(versionString string) error {
	rootCmd.Version = versionString
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.json", "Pfad zur Konfigurationsdatei")

	rootCmd.AddCommand(inviteCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(setupCmd)
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
