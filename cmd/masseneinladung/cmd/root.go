package cmd

import (
	"fmt"
	"os"

	"github.com/janmz/masseneinladung/internal/version"
	"github.com/spf13/cobra"
)

var configPath string

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "masseneinladung",
	Short: "ChurchTools-Einladungen aus CSV versenden",
	Long: `Masseneinladung liest Personen-IDs aus einer CSV-Datei und
versendet über die ChurchTools-API Einladungs-E-Mails.

Nutze 'setup' für Ersteinrichtung von URL, Login-Token und Berechtigungsprüfung.`,
	Version: fmt.Sprintf("%s (build %d, %s)", version.Version, version.Build, version.BuildTime),
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
