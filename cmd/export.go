package cmd

import (
	"fmt"
	"os"
	"strings"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	config "github.com/janmz/churchtools-invite/internal/config"
	csvfile "github.com/janmz/churchtools-invite/internal/csvfile"

	"github.com/spf13/cobra"
)

var (
	exportOutput          string
	exportGroup           string
	exportCampusFlag      string
	exportStatus          string
	exportInteractive     bool
	exportInvited         bool
	exportAllCampuses     bool
	exportSkipPermRequest bool
	exportSkipPreJoin     bool
	exportForce           bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Personenliste im CSV-Format für Einladungen exportieren",
	Long: `Lädt Personen aus ChurchTools und schreibt eine CSV-Datei im Format
id,vorname,nachname,email,standort,status – direkt verwendbar für invite
(Spalten standort und status werden beim Import ignoriert).

Standardmäßig werden nur Personen exportiert, die noch nicht eingeladen wurden
(status NEU). Mit --invited erscheinen auch Eingeladene und Registrierte.

Mit --interactive wählen Sie Standort, optional einen Filter (Personenstatus
oder Gruppe) und den Einladungsstatus (Neu, Eingeladen, Registriert).

--status filtert nach dem Personenstatus (z. B. Mitglied, Gast), --group nach
einer Gruppe – jeweils per ID oder Name. Welche Werte es in Ihrer Gemeinde
gibt, zeigen "list statuses" und "list groups". Nicht zu verwechseln mit der
CSV-Spalte status (Einladungsstatus NEU/Eingeladen/Registriert).

Eine vorhandene Ziel-Datei wird nur mit --force überschrieben.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExport(cmd)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "personen.csv", "Ziel-Datei (- für stdout)")
	exportCmd.Flags().StringVar(&exportCampusFlag, "campus", "", "Standort-ID, eindeutiger Namens-Teilstring oder \"all\"")
	exportCmd.Flags().StringVar(&exportStatus, "status", "", "Nur Personen mit diesem Personenstatus: ID oder Name, z. B. Mitglied (Übersicht: list statuses)")
	exportCmd.Flags().StringVar(&exportGroup, "group", "", "Nur Mitglieder dieser Gruppe: ID oder Name (Übersicht: list groups)")
	// Former names, kept so existing scripts keep working.
	exportCmd.Flags().StringVar(&exportStatus, "status-id", "", "Alias für --status")
	exportCmd.Flags().StringVar(&exportGroup, "group-id", "", "Alias für --group")
	_ = exportCmd.Flags().MarkHidden("status-id")
	_ = exportCmd.Flags().MarkHidden("group-id")
	exportCmd.Flags().BoolVar(&exportInteractive, "interactive", false, "Standort und Filter interaktiv auswählen")
	exportCmd.Flags().BoolVarP(&exportInvited, "invited", "i", false, "Auch bereits eingeladene und registrierte Personen exportieren")
	exportCmd.Flags().BoolVar(&exportAllCampuses, "all-campuses", false, "Keinen Standort-Filter anwenden (Alias für --campus all)")
	exportCmd.Flags().BoolVar(&exportSkipPermRequest, "skip-permission-request", false, "Keine Gruppenmitgliedschaft für fehlende Berechtigungen beantragen")
	exportCmd.Flags().BoolVar(&exportSkipPreJoin, "skip-pre-join-groups", false, "Keine Vorab-Gruppen vor dem Export beitreten")
	exportCmd.Flags().BoolVar(&exportForce, "force", false, "Vorhandene Ziel-Datei überschreiben")
}

func runExport(cmd *cobra.Command) error {
	if err := requireNonEmptyFlag(cmd, "--output (-o)", exportOutput); err != nil {
		return err
	}
	if err := validatePathFlagValueForCmd(cmd, "--output", exportOutput); err != nil {
		return err
	}
	// Check before the export so an edited CSV is not overwritten.
	if exportOutput != "-" && !exportForce {
		if _, err := os.Stat(exportOutput); err == nil {
			return fmt.Errorf("%s existiert bereits – anderen Dateinamen wählen oder --force zum Überschreiben", exportOutput)
		}
	}

	cfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}

	client, err := connectChurchTools(cfg)
	if err != nil {
		return err
	}
	// Leave groups joined only for this run, also when the run fails.
	defer leaveTemporaryGroups(client)

	if !exportSkipPreJoin {
		if err := ensurePreJoinGroups(client, cfg); err != nil {
			return err
		}
	}

	if !exportSkipPermRequest {
		if err := ensureExportPermissions(client); err != nil {
			return err
		}
	}

	var opts churchtools.PersonListOptions
	inviteFilter := exportInviteFilterNEU
	if exportInvited {
		inviteFilter = exportInviteFilterAll
	}
	if exportInteractive {
		opts, inviteFilter, err = interactiveExportOptions(client, &cfg)
		if err != nil {
			return err
		}
	} else {
		statusID, groupID, err := resolveStatusAndGroup(client, exportStatus, exportGroup)
		if err != nil {
			return err
		}
		opts = churchtools.PersonListOptions{
			GroupID:  groupID,
			StatusID: statusID,
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
	if err := interrupted(); err != nil {
		return err
	}
	persons = filterExportPersons(persons, inviteFilter)
	if len(persons) == 0 {
		return fmt.Errorf("keine Personen gefunden (%s)", describeExportFilters(opts, inviteFilter))
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
		fmt.Fprintf(os.Stderr, "%d Personen nach stdout exportiert (%s)\n", len(entries), describeExportFilters(opts, inviteFilter))
		return nil
	}

	if err := csvfile.Write(exportOutput, entries); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%d Personen nach %s exportiert (%s)\n", len(entries), exportOutput, describeExportFilters(opts, inviteFilter))
	printDeletionHint(exportOutput)
	return nil
}

// resolveStatusAndGroup turns the --status and --group values (ID or name)
// into IDs; empty values mean "no filter".
func resolveStatusAndGroup(client *churchtools.Client, status, group string) (int, int, error) {
	var statusID, groupID int
	if strings.TrimSpace(status) != "" {
		items, err := statusItems(client)
		if err != nil {
			return 0, 0, err
		}
		if statusID, err = resolveNamedItem(statusKind, items, status); err != nil {
			return 0, 0, err
		}
	}
	if strings.TrimSpace(group) != "" {
		items, err := groupItems(client, 0)
		if err != nil {
			return 0, 0, err
		}
		if groupID, err = resolveNamedItem(groupKind, items, group); err != nil {
			return 0, 0, err
		}
	}
	return statusID, groupID, nil
}

// printDeletionHint reminds the user that the CSV holds personal data
// (GDPR storage limitation).
func printDeletionHint(path string) {
	fmt.Fprintf(os.Stderr,
		"Datenschutz: %s enthält personenbezogene Daten. Datei nach Abschluss der Einladungen löschen.\n", path)
}

type exportInviteFilter string

const (
	exportInviteFilterAll         exportInviteFilter = ""
	exportInviteFilterNEU         exportInviteFilter = "NEU"
	exportInviteFilterEingeladen  exportInviteFilter = "Eingeladen"
	exportInviteFilterRegistriert exportInviteFilter = "Registriert"
)

func filterExportPersons(persons []churchtools.Person, filter exportInviteFilter) []churchtools.Person {
	if filter == exportInviteFilterAll {
		return persons
	}
	filtered := make([]churchtools.Person, 0, len(persons))
	for _, person := range persons {
		if person.ExportStatusLabel() == string(filter) {
			filtered = append(filtered, person)
		}
	}
	return filtered
}

func describeInviteFilter(filter exportInviteFilter) string {
	switch filter {
	case exportInviteFilterAll:
		return "alle Einladungsstatus"
	case exportInviteFilterEingeladen:
		return "nur Eingeladen"
	case exportInviteFilterRegistriert:
		return "nur Registriert"
	default:
		return "nur NEU"
	}
}
