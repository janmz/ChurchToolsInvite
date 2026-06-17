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
	exportOutput      string
	exportGroup       int
	exportCampus      int
	exportStatus      int
	exportInteractive bool
	exportAllCampuses       bool
	exportSkipPermRequest   bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Personenliste im CSV-Format für Einladungen exportieren",
	Long: `Lädt Personen aus ChurchTools und schreibt eine CSV-Datei im Format
id,vorname,nachname,email – direkt verwendbar für invite.

Mit --interactive wählen Sie zuerst einen Standort und danach optional
einen Filter (alle Personen, Personenstatus oder Gruppe).`,
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(runExport())
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "personen.csv", "Ziel-Datei (- für stdout)")
	exportCmd.Flags().IntVar(&exportCampus, "campus-id", 0, "Nur Personen dieses Standorts exportieren")
	exportCmd.Flags().IntVar(&exportStatus, "status-id", 0, "Nur Personen mit diesem Status exportieren")
	exportCmd.Flags().IntVar(&exportGroup, "group-id", 0, "Nur Mitglieder dieser Gruppe exportieren")
	exportCmd.Flags().BoolVarP(&exportInteractive, "interactive", "i", false, "Standort und Filter interaktiv auswählen")
	exportCmd.Flags().BoolVar(&exportAllCampuses, "all-campuses", false, "Keinen Standort-Filter anwenden (Standard: Standort des angemeldeten Nutzers)")
	exportCmd.Flags().BoolVar(&exportSkipPermRequest, "skip-permission-request", false, "Keine Gruppenmitgliedschaft für fehlende Berechtigungen beantragen")
}

func runExport() error {
	if exportOutput == "" {
		return fmt.Errorf("--output ist erforderlich")
	}

	cfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}

	client, err := connectChurchTools(cfg)
	if err != nil {
		return err
	}

	if !exportSkipPermRequest {
		if err := ensureExportPermissions(client, cfg); err != nil {
			return err
		}
	}

	var opts churchtools.PersonListOptions
	if exportInteractive {
		opts, err = interactiveExportOptions(client, &cfg)
		if err != nil {
			return err
		}
	} else {
		opts = churchtools.PersonListOptions{
			GroupID:  exportGroup,
			CampusID: exportCampus,
			StatusID: exportStatus,
		}
		if !exportAllCampuses {
			if err := applyDefaultCampus(client, &cfg, &opts); err != nil {
				return err
			}
		}
	}

	persons, err := client.ListPersons(opts)
	if err != nil {
		return err
	}
	if len(persons) == 0 {
		return fmt.Errorf("keine personen gefunden (%s)", describeExportFilters(opts))
	}

	entries := csvfile.EntriesFromPersons(persons)

	if exportOutput == "-" {
		if err := csvfile.WriteTo(os.Stdout, entries); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "%d personen nach stdout exportiert (%s)\n", len(entries), describeExportFilters(opts))
		return nil
	}

	if err := csvfile.Write(exportOutput, entries); err != nil {
		return err
	}

	fmt.Printf("%d personen nach %s exportiert (%s)\n", len(entries), exportOutput, describeExportFilters(opts))
	return nil
}
