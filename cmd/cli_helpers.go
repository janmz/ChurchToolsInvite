package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func validatePathFlagValueForCmd(cmd *cobra.Command, flagName, value string) error {
	if err := validatePathFlagValue(flagName, value); err != nil {
		return usageErrorf(cmd, "%s", err.Error())
	}
	return nil
}

func requireNonEmptyFlag(cmd *cobra.Command, flagLabel, value string) error {
	if value == "" {
		return usageErrorf(cmd, "Pflichtoption %s fehlt", flagLabel)
	}
	return nil
}

func applyGermanFlagLabels(cmd *cobra.Command) {
	cmd.InitDefaultHelpFlag()
	cmd.InitDefaultVersionFlag()
	if f := cmd.Flags().Lookup("help"); f != nil {
		f.Usage = "Hilfe anzeigen"
	}
	if f := cmd.Flags().Lookup("version"); f != nil {
		f.Usage = "Version anzeigen"
	}
	for _, sub := range cmd.Commands() {
		if sub.Name() == "help" {
			sub.Short = "Hilfe zu einem Befehl anzeigen"
			sub.Long = fmt.Sprintf(
				"Hilfe zu einem Befehl anzeigen.\n\nAufruf: %s help [Befehl]",
				cmd.DisplayName(),
			)
		}
		applyGermanFlagLabels(sub)
	}
}

func setupSubcommandArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		applyGermanFlagLabels(cmd)
		_ = cmd.Usage()
		return fmt.Errorf("unbekannter Unterbefehl %q", args[0])
	}
	return nil
}

func setupSubcommandRequired(cmd *cobra.Command, args []string) error {
	applyGermanFlagLabels(cmd)
	_ = cmd.Usage()
	return fmt.Errorf("bitte einen Unterbefehl angeben: init, test, token, permissions")
}
