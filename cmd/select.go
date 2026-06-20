package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type menuItem struct {
	id   int
	name string
}

func promptMenu(title string, items []menuItem, allowSkip bool) (int, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("keine einträge für %s", title)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n%s:\n", title)
	for i, item := range items {
		fmt.Printf("  [%d] %s (ID %d)\n", i+1, item.name, item.id)
	}
	if allowSkip {
		fmt.Println("  [0] Kein zusätzlicher Filter")
	}
	fmt.Print("Auswahl: ")

	line, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	choice, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return 0, fmt.Errorf("ungültige auswahl")
	}
	if allowSkip && choice == 0 {
		return 0, nil
	}
	if choice < 1 || choice > len(items) {
		return 0, fmt.Errorf("auswahl außerhalb des gültigen bereichs")
	}
	return items[choice-1].id, nil
}

func promptExportInviteStatus() (exportInviteFilter, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\nEinladungsstatus:")
	fmt.Println("  [n] Neu (noch nicht eingeladen)")
	fmt.Println("  [e] Eingeladen")
	fmt.Println("  [r] Registriert")
	fmt.Print("Auswahl [n/e/r]: ")

	line, err := reader.ReadString('\n')
	if err != nil {
		return exportInviteFilterNEU, err
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "n", "neu":
		return exportInviteFilterNEU, nil
	case "e", "eingeladen":
		return exportInviteFilterEingeladen, nil
	case "r", "registriert":
		return exportInviteFilterRegistriert, nil
	default:
		return exportInviteFilterNEU, fmt.Errorf("ungültige einladungsstatus-auswahl")
	}
}

func promptFilterMode() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\nZusätzlicher Filter:")
	fmt.Println("  [a] Alle Personen am Standort")
	fmt.Println("  [s] Nach Personenstatus filtern")
	fmt.Println("  [g] Nach Gruppe filtern")
	fmt.Print("Auswahl [a/s/g]: ")

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "a", "alle":
		return "all", nil
	case "s", "status":
		return "status", nil
	case "g", "gruppe", "group":
		return "group", nil
	default:
		return "", fmt.Errorf("ungültige filterauswahl")
	}
}
