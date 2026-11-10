package cmd

import (
	"fmt"
	"os"
	"strings"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	config "github.com/janmz/churchtools-invite/internal/config"
	"github.com/janmz/churchtools-invite/internal/termio"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "ChurchTools-Verbindung und Berechtigungen einrichten",
	Args:  setupSubcommandArgs,
	RunE:  setupSubcommandRequired,
}

var setupInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interaktive Erstellung der config.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetupInit()
	},
}

var setupTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Verbindung und Login testen",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetupTest()
	},
}

var setupTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Eigenen Login-Token abrufen und verschlüsselt in config.json speichern",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetupToken()
	},
}

var setupPermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Berechtigungen für Einladungen prüfen",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetupPermissions()
	},
}

func init() {
	setupCmd.AddCommand(setupInitCmd)
	setupCmd.AddCommand(setupTestCmd)
	setupCmd.AddCommand(setupTokenCmd)
	setupCmd.AddCommand(setupPermissionsCmd)
}

// runSetupInit asks for connection settings. An existing config file is
// loaded first (after confirmation) and only the asked fields are replaced,
// so campus_id, delay_ms and other settings survive.
func runSetupInit() error {
	cfg := config.Default()
	if config.Exists(configPath) {
		answer, err := readLine(fmt.Sprintf("%s existiert bereits. Verbindungsdaten neu eingeben (übrige Einstellungen bleiben erhalten)? [j/N]: ", configPath))
		if err != nil {
			return err
		}
		if !strings.EqualFold(strings.TrimSpace(answer), "j") {
			return fmt.Errorf("abgebrochen – %s unverändert", configPath)
		}
		existing, err := config.LoadFile(configPath)
		if err != nil {
			return err
		}
		cfg = existing
	}

	instanceName, err := readLine("ChurchTools-Instanz (z. B. meine-gemeinde oder meine-gemeinde-unterstadt): ")
	if err != nil {
		return err
	}
	baseURL, err := config.BaseURLFromInstanceName(instanceName)
	if err != nil {
		return err
	}
	cfg.BaseURL = baseURL

	mainURL, err := askMainInstanceURL(baseURL, cfg.MainURL)
	if err != nil {
		return err
	}
	cfg.MainURL = mainURL

	method, err := readLine("Login-Methode [token/password]: ")
	if err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(method)) {
	case "password", "passwort", "pw":
		user, err := readLine("Benutzername: ")
		if err != nil {
			return err
		}
		pass, err := termio.ReadPassword("Passwort: ", stdinReader)
		if err != nil {
			return fmt.Errorf("passwort einlesen: %w", err)
		}
		// The password is stored verbatim; spaces may be part of it.
		cfg.UsePassword(user, pass)
	default:
		// Read hidden like a password: the token grants full account access.
		token, err := termio.ReadPassword("Login-Token: ", stdinReader)
		if err != nil {
			return fmt.Errorf("Login-Token einlesen: %w", err)
		}
		cfg.UseLoginToken(token)
	}

	client, err := connectChurchTools(cfg)
	if err != nil {
		return fmt.Errorf("verbindungstest fehlgeschlagen: %w", err)
	}

	user, err := client.WhoAmI()
	if err != nil {
		return err
	}

	fmt.Printf("Login erfolgreich als %s %s (%s)\n", user.FirstName, user.LastName, user.Email)

	if cfg.LoginTokenPassword == "" {
		// Ask first: ChurchTools creates a long-lived token on request
		// ("get or create"), which should only happen when it is wanted.
		answer, _ := readLine(fmt.Sprintf("\nOptional: Login-Token abrufen und verschlüsselt in %s speichern (statt Passwort)? [j/N]: ", configPath))
		if strings.EqualFold(strings.TrimSpace(answer), "j") {
			token, err := client.MeAPIToken()
			if err != nil || token == "" {
				token, err = client.LoginToken(user.ID)
			}
			if err == nil && token != "" {
				cfg.UseLoginToken(token)
			} else {
				fmt.Fprintln(os.Stderr, "Login-Token konnte nicht abgerufen werden – Passwort bleibt gespeichert.")
			}
		}
	}

	// An empty answer keeps the previous value; an empty value means the
	// built-in default, applied at runtime (not written to the file).
	shownPreJoin := cfg.PreJoinGroups
	if strings.TrimSpace(shownPreJoin) == "" {
		shownPreJoin = config.DefaultPreJoinGroups + " (Standard)"
	}
	preJoinInput, _ := readLine(fmt.Sprintf("Gruppen vor Export/Invite (kommagetrennt, - = keine) [%s]: ", shownPreJoin))
	if preJoinInput = strings.TrimSpace(preJoinInput); preJoinInput != "" {
		cfg.PreJoinGroups = preJoinInput
	}

	if err := config.Save(configPath, cfg); err != nil {
		return err
	}

	fmt.Printf("Konfiguration gespeichert: %s\n", configPath)
	return nil
}

func runSetupTest() error {
	cfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}

	client, err := connectChurchTools(cfg)
	if err != nil {
		return fmt.Errorf("login fehlgeschlagen: %w", err)
	}
	if err := client.Ping(); err != nil {
		return err
	}

	user, err := client.WhoAmI()
	if err != nil {
		return err
	}

	fmt.Println("Verbindung OK")
	fmt.Printf("Angemeldet als: %s %s (%s)\n", user.FirstName, user.LastName, user.Email)
	return nil
}

// runSetupToken fetches the login token of the logged-in user and stores it
// encrypted in the config file. The token is never printed, and tokens of
// other persons cannot be requested (that would take over their account).
func runSetupToken() error {
	cfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}

	client, err := connectChurchTools(cfg)
	if err != nil {
		return err
	}

	user, err := client.WhoAmI()
	if err != nil {
		return err
	}
	token, err := client.LoginToken(user.ID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("ChurchTools hat keinen Login-Token geliefert")
	}

	fmt.Printf("Login-Token für %s %s (Person %d) abgerufen.\n", user.FirstName, user.LastName, user.ID)
	answer, _ := readLine(fmt.Sprintf("Token verschlüsselt in %s speichern (Benutzername/Passwort werden dort entfernt)? [j/N]: ", configPath))
	if !strings.EqualFold(strings.TrimSpace(answer), "j") {
		fmt.Println("Nicht gespeichert.")
		return nil
	}

	// Load the file without environment values, so nothing from CT_* is
	// written to disk.
	fileCfg := config.Default()
	if config.Exists(configPath) {
		if fileCfg, err = config.LoadFile(configPath); err != nil {
			return err
		}
	} else {
		fileCfg.BaseURL = cfg.BaseURL
		fileCfg.MainURL = cfg.MainURL
	}
	fileCfg.UseLoginToken(token)
	if err := config.Save(configPath, fileCfg); err != nil {
		return err
	}
	fmt.Printf("Login-Token verschlüsselt in %s gespeichert.\n", configPath)
	return nil
}

// askMainInstanceURL asks for the main (central) instance of a sub-instance.
// A suggestion derived from the instance name is shown with its host, but
// only used after explicit confirmation: the derived host may belong to a
// different church, and credentials are sent there on login fallback.
func askMainInstanceURL(baseURL, current string) (string, error) {
	// Preferred: the main instance the sub-instance itself redirects its
	// OAuth login to. A host derived from the name is only a guess (it may
	// belong to another church) and must be typed in, not confirmed with "j".
	suggestion, err := discoverMainInstanceURL(baseURL)
	ok := err == nil && suggestion != ""
	guess, guessOK := churchtools.MainInstanceURL(baseURL)
	if !ok && !guessOK && current == "" {
		return "", nil
	}

	fmt.Fprintln(os.Stderr, "\nNebeninstanz: Schlägt die Anmeldung fehl, kann sich das Programm über die")
	fmt.Fprintln(os.Stderr, "Hauptinstanz eurer Gemeinde anmelden. Zugangsdaten gehen dann an diese Adresse.")
	switch {
	case ok:
		fmt.Fprintf(os.Stderr, "Die Nebeninstanz meldet als Hauptinstanz: %s\n", suggestion)
	case guessOK:
		fmt.Fprintf(os.Stderr, "Vermutung aus dem Namen (nicht bestätigt, ggf. andere Gemeinde): %s – zum Übernehmen die URL eintippen.\n", guess)
	}
	prompt := "Hauptinstanz-URL (leer = keine"
	if ok {
		prompt = "Hauptinstanz-URL ([j] = gemeldete übernehmen, leer = keine"
	}
	if current != "" {
		prompt += ", [=] = " + current + " behalten"
	}
	answer, err := readLine(prompt + "): ")
	if err != nil {
		return "", err
	}

	answer = strings.TrimSpace(answer)
	switch strings.ToLower(answer) {
	case "", "n", "nein", "no", "-":
		return "", nil
	case "=":
		if current != "" {
			return current, nil
		}
	case "j", "ja", "y", "yes":
		if ok {
			return suggestion, nil
		}
		return "", fmt.Errorf("keine von der Nebeninstanz gemeldete Hauptinstanz – URL eingeben oder leer lassen")
	}

	// Only a full URL or an instance name with at least two characters is
	// accepted, so a typo like "x" does not become https://x.church.tools.
	if len(answer) < 2 {
		return "", fmt.Errorf("ungültige Eingabe %q für die Hauptinstanz", answer)
	}
	mainURL := config.NormalizeBaseURL(answer)
	if err := config.ValidateInstanceURL("main_url", mainURL); err != nil {
		return "", err
	}
	if strings.EqualFold(mainURL, config.NormalizeBaseURL(baseURL)) {
		return "", fmt.Errorf("Hauptinstanz %s ist identisch mit der konfigurierten Instanz", mainURL)
	}
	if ok && !strings.EqualFold(mainURL, suggestion) {
		fmt.Fprintf(os.Stderr, "Warnung: %s weicht von der gemeldeten Hauptinstanz %s ab – die Anmeldung über OAuth wird dann abgelehnt.\n", mainURL, suggestion)
	}
	return mainURL, nil
}

// discoverMainInstanceURL is replaceable in tests (no network access).
var discoverMainInstanceURL = churchtools.DiscoverMainInstanceURL

func runSetupPermissions() error {
	cfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}

	client, err := connectChurchTools(cfg)
	if err != nil {
		return err
	}

	perms, err := client.GlobalPermissions()
	if err != nil {
		return fmt.Errorf("berechtigungen konnten nicht geladen werden: %w", err)
	}

	found := churchtools.FindInvitePermissions(perms)
	fmt.Println("Benötigte Berechtigungen (laut ChurchTools-Dokumentation):")
	for _, hint := range churchtools.PermissionHints {
		fmt.Printf("  - %s\n", hint)
	}

	fmt.Println("\nGefundene invite-/einlad-bezogene Einträge in /permissions/global:")
	if len(found) == 0 {
		fmt.Println("  (keine expliziten Treffer – prüfe Rechte manuell in ChurchTools)")
	} else {
		for _, item := range found {
			fmt.Printf("  - %s\n", item)
		}
	}

	fmt.Println("\nHinweis: Für setup token wird zusätzlich die Berechtigung zum Lesen")
	fmt.Println("des Login-Tokens benötigt (Profil → Berechtigungen / API-Token).")
	return nil
}
