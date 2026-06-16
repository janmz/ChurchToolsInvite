package cmd

import (
	"fmt"
	"os"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	config "github.com/janmz/churchtools-invite/internal/config"
	csvfile "github.com/janmz/churchtools-invite/internal/csvfile"

	"github.com/spf13/cobra"
)

var (
	exportOutput string
	exportGroup  int
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Personenliste im CSV-Format für Einladungen exportieren",
	Long: `Lädt Personen aus ChurchTools und schreibt eine CSV-Datei im Format
id,vorname,nachname,email – direkt verwendbar für invite und validate.`,
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(runExport())
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "personen.csv", "Ziel-Datei (- für stdout)")
	exportCmd.Flags().IntVar(&exportGroup, "group-id", 0, "Nur Mitglieder dieser Gruppe exportieren")
}

func runExport() error {
	if exportOutput == "" {
		return fmt.Errorf("--output ist erforderlich")
	}

	cfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}

	client := churchtools.NewClient(cfg.BaseURL, cfg.LoginToken, cfg.Username, cfg.Password)
	if err := client.Login(); err != nil {
		return err
	}

	opts := churchtools.PersonListOptions{GroupID: exportGroup}
	persons, err := client.ListPersons(opts)
	if err != nil {
		return err
	}
	if len(persons) == 0 {
		return fmt.Errorf("keine personen gefunden")
	}

	entries := csvfile.EntriesFromPersons(persons)

	if exportOutput == "-" {
		if err := csvfile.WriteTo(os.Stdout, entries); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "%d personen nach stdout exportiert\n", len(entries))
		return nil
	}

	if err := csvfile.Write(exportOutput, entries); err != nil {
		return err
	}

	fmt.Printf("%d personen nach %s exportiert\n", len(entries), exportOutput)
	return nil
}
