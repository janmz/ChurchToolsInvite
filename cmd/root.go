package cmd

import (
	"github.com/spf13/cobra"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "Churchtools-Invite",
	Short: "ChurchTools-Einladungen aus CSV versenden",
	Long: `ChurchTools-Invite liest Personen-IDs aus einer CSV-Datei und
versendet über die ChurchTools-API Einladungs-E-Mails.

Nutze 'setup init' für Ersteinrichtung von URL, Login-Token und Berechtigungsprüfung.`,
	Version: "undefined",
}

// Execute runs the root command.
func Execute(versionString string) error {
	rootCmd.Version = versionString
	rootCmd.InitDefaultHelpCmd()
	applyGermanFlagLabels(rootCmd)
	return finalizeCLIError(rootCmd.Execute())
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetUsageTemplate(usageTemplateDE)
	rootCmd.SetHelpTemplate(helpTemplateDE)
	rootCmd.SetVersionTemplate("{{.Name}} Version {{.Version}}\n")

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.json", "Pfad zur Konfigurationsdatei")

	rootCmd.AddCommand(inviteCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(setupCmd)

	configureGermanCLI(rootCmd)
}
