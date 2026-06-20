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
	exportOutput            string
	exportGroup             int
	exportCampusFlag        string
	exportStatus            int
	exportInteractive       bool
	exportInvited           bool
	exportAllCampuses       bool
	exportSkipPermRequest   bool
	exportSkipPreJoin       bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Personenliste im CSV-Format für Einladungen exportieren",
	Long: `Lädt Personen aus ChurchTools und schreibt eine CSV-Datei im Format
id,vorname,nachname,email,standort,status – direkt verwendbar für invite
(Spalten standort und status werden beim Import ignoriert).

Standardmäßig werden nur Personen exportiert, die noch nicht eingeladen wurden
(status NEU). Mit --invited erscheinen auch Eingeladene und Registrierte.

Mit --interactive wählen Sie zuerst einen Standort und danach optional
einen Filter (alle Personen, Personenstatus oder Gruppe).`,
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(runExport())
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "personen.csv", "Ziel-Datei (- für stdout)")
	exportCmd.Flags().StringVar(&exportCampusFlag, "campus", "", "Standort-ID, eindeutiger Namens-Teilstring oder \"all\"")
	exportCmd.Flags().IntVar(&exportStatus, "status-id", 0, "Nur Personen mit diesem Status exportieren")
	exportCmd.Flags().IntVar(&exportGroup, "group-id", 0, "Nur Mitglieder dieser Gruppe exportieren")
	exportCmd.Flags().BoolVar(&exportInteractive, "interactive", false, "Standort und Filter interaktiv auswählen")
	exportCmd.Flags().BoolVarP(&exportInvited, "invited", "i", false, "Auch bereits eingeladene und registrierte Personen exportieren")
	exportCmd.Flags().BoolVar(&exportAllCampuses, "all-campuses", false, "Keinen Standort-Filter anwenden (Alias für --campus all)")
	exportCmd.Flags().BoolVar(&exportSkipPermRequest, "skip-permission-request", false, "Keine Gruppenmitgliedschaft für fehlende Berechtigungen beantragen")
	exportCmd.Flags().BoolVar(&exportSkipPreJoin, "skip-pre-join-groups", false, "Keine Vorab-Gruppen vor dem Export beitreten")
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

	if !exportSkipPreJoin {
		if err := ensurePreJoinGroups(client, cfg); err != nil {
			return err
		}
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
			StatusID: exportStatus,
		}
		choice, err := resolveNonInteractiveCampus(client, &cfg, exportCampusFlag, exportAllCampuses)
		if err != nil {
			return err
		}
		applyCampusChoice(&opts, choice)
	}

	persons, err := client.ListPersons(opts)
	if err != nil {
		return err
	}
	persons = filterExportPersons(persons, exportInvited)
	if len(persons) == 0 {
		return fmt.Errorf("keine personen gefunden (%s)", describeExportFilters(opts, exportInvited))
	}

	campusNames, err := campusNameMap(client)
	if err != nil {
		return err
	}

	entries := csvfile.EntriesFromPersons(persons, campusNames)

	if exportOutput == "-" {
		if err := csvfile.WriteTo(os.Stdout, entries); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "%d personen nach stdout exportiert (%s)\n", len(entries), describeExportFilters(opts, exportInvited))
		return nil
	}

	if err := csvfile.Write(exportOutput, entries); err != nil {
		return err
	}

	fmt.Printf("%d personen nach %s exportiert (%s)\n", len(entries), exportOutput, describeExportFilters(opts, exportInvited))
	return nil
}

func filterExportPersons(persons []churchtools.Person, includeInvited bool) []churchtools.Person {
	if includeInvited {
		return persons
	}
	filtered := make([]churchtools.Person, 0, len(persons))
	for _, person := range persons {
		if person.ExportStatusLabel() == "NEU" {
			filtered = append(filtered, person)
		}
	}
	return filtered
}
