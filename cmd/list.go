package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	config "github.com/janmz/churchtools-invite/internal/config"
	"github.com/janmz/churchtools-invite/internal/textnorm"
	"github.com/spf13/cobra"
)

var (
	statusKind = lookupKind{flag: "--status", label: "Personenstatus", listCmd: "list statuses"}
	groupKind  = lookupKind{flag: "--group", label: "Gruppe", listCmd: "list groups"}
	campusKind = lookupKind{flag: "--campus", label: "Standort", listCmd: "list campuses"}
)

var (
	listGroupsCampus string
	listGroupsSearch string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Personenstatus, Gruppen oder Standorte mit ID und Name anzeigen",
	Long: `Zeigt die Werte, die bei export --status, --group und --campus angegeben
werden können. Statt der ID darf dort auch der Name (oder ein eindeutiger
Teil davon) stehen. Status, Gruppen und Standorte legt jede Gemeinde selbst
an, die IDs unterscheiden sich daher je ChurchTools-Instanz.`,
	Args: setupSubcommandArgs,
	RunE: setupSubcommandRequired,
}

var listStatusesCmd = &cobra.Command{
	Use:   "statuses",
	Short: "Personenstatus (Mitglied, Gast, …) mit ID anzeigen – für export --status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList(func(client *churchtools.Client) ([]namedItem, error) { return statusItems(client) })
	},
}

var listGroupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Gruppen mit ID anzeigen – für export --group",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList(func(client *churchtools.Client) ([]namedItem, error) {
			campusID := 0
			if strings.TrimSpace(listGroupsCampus) != "" {
				choice, err := parseExportCampusFlag(client, listGroupsCampus)
				if err != nil {
					return nil, err
				}
				campusID = choice.campusID
			}
			items, err := groupItems(client, campusID)
			if err != nil {
				return nil, err
			}
			return filterItems(items, listGroupsSearch), nil
		})
	},
}

var listCampusesCmd = &cobra.Command{
	Use:   "campuses",
	Short: "Standorte mit ID anzeigen – für export --campus",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList(func(client *churchtools.Client) ([]namedItem, error) { return campusItems(client) })
	},
}

func init() {
	listGroupsCmd.Flags().StringVar(&listGroupsCampus, "campus", "", "Nur Gruppen dieses Standorts (ID, Name oder \"all\")")
	listGroupsCmd.Flags().StringVar(&listGroupsSearch, "search", "", "Nur Gruppen, deren Name diesen Text enthält")
	listCmd.AddCommand(listStatusesCmd, listGroupsCmd, listCampusesCmd)
	rootCmd.AddCommand(listCmd)
}

// runList connects, loads the items and prints them as "ID  Name", sorted
// by name. Only read access is used; no groups are joined.
func runList(load func(*churchtools.Client) ([]namedItem, error)) error {
	cfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}
	client, err := connectChurchTools(cfg)
	if err != nil {
		return err
	}
	items, err := load(client)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Fprintln(os.Stderr, "Keine Einträge sichtbar (fehlende Berechtigung? ggf. Vorab-Gruppen per export/invite beitreten).")
		return nil
	}
	sort.Slice(items, func(i, j int) bool {
		left, right := strings.ToLower(items[i].Name), strings.ToLower(items[j].Name)
		if left != right {
			return left < right
		}
		return items[i].ID < items[j].ID
	})
	width := 1
	for _, item := range items {
		if w := len(fmt.Sprint(item.ID)); w > width {
			width = w
		}
	}
	for _, item := range items {
		fmt.Printf("%*d  %s\n", width, item.ID, textnorm.Printable(item.Name))
	}
	return nil
}

func filterItems(items []namedItem, search string) []namedItem {
	needle := textnorm.FoldAlnum(search)
	if needle == "" {
		return items
	}
	var filtered []namedItem
	for _, item := range items {
		if strings.Contains(textnorm.FoldAlnum(item.Name), needle) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func statusItems(client *churchtools.Client) ([]namedItem, error) {
	statuses, err := client.ListPersonStatuses()
	if err != nil {
		return nil, err
	}
	items := make([]namedItem, len(statuses))
	for i, status := range statuses {
		items[i] = namedItem{ID: status.ID, Name: status.Name}
	}
	return items, nil
}

func groupItems(client *churchtools.Client, campusID int) ([]namedItem, error) {
	groups, err := client.ListGroups(churchtools.GroupListOptions{CampusID: campusID})
	if err != nil {
		return nil, err
	}
	items := make([]namedItem, len(groups))
	for i, group := range groups {
		items[i] = namedItem{ID: group.ID, Name: churchtools.PlainGroupName(group.Name)}
	}
	return items, nil
}

func campusItems(client *churchtools.Client) ([]namedItem, error) {
	campuses, err := client.ListCampuses()
	if err != nil {
		return nil, err
	}
	return campusesToItems(campuses), nil
}

func campusesToItems(campuses []churchtools.Campus) []namedItem {
	items := make([]namedItem, len(campuses))
	for i, campus := range campuses {
		items[i] = namedItem{ID: campus.ID, Name: campus.Name}
	}
	return items
}
