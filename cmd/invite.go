package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	config "github.com/janmz/churchtools-invite/internal/config"
	csvfile "github.com/janmz/churchtools-invite/internal/csvfile"
	invite "github.com/janmz/churchtools-invite/internal/invite"

	"github.com/spf13/cobra"
)

var (
	csvPath               string
	dryRun                bool
	delayMS               int
	noSyncEmail           bool
	skipPermissionRequest bool
	skipPreJoin           bool
	reinvite              bool
	assumeYes             bool
	ignoreNameMismatch    bool
	allowSingleEmail      bool
)

var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Einladungen für alle Personen in der CSV senden",
	Long: `Liest Personen-IDs aus einer CSV-Datei und versendet Einladungs-E-Mails
über die ChurchTools-API.

Pflichtoption: --csv (-f) mit Pfad zur CSV-Datei.

Optionen wie --dry-run, --reinvite und --no-sync-email steuern Ablauf und
E-Mail-Abgleich; siehe Optionen unten.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInvite(cmd)
	},
}

func init() {
	inviteCmd.Flags().StringVarP(&csvPath, "csv", "f", "", "Pfad zur CSV-Datei (Pflicht)")
	inviteCmd.Flags().IntVar(&delayMS, "delay-ms", 0, "Pause nach jeder Einladung/E-Mail-Änderung in Millisekunden, 0–60000 (0 = config.delay_ms)")
	inviteCmd.Flags().BoolVar(&skipPermissionRequest, "skip-permission-request", false, "Keine Gruppenmitgliedschaft für fehlende Berechtigungen beantragen")
	inviteCmd.Flags().BoolVar(&skipPreJoin, "skip-pre-join-groups", false, "Keine Vorab-Gruppen vor dem Invite beitreten")
	inviteCmd.Flags().BoolVar(&reinvite, "reinvite", false, "Bereits eingeladene (Status Eingeladen) erneut einladen; Registrierte werden immer übersprungen")
	inviteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Prüfen/simulieren ohne Einladungen zu senden oder Personen zu ändern (Vorab-Gruppen werden für Leserechte beigetreten und danach wieder verlassen)")
	inviteCmd.Flags().BoolVar(&noSyncEmail, "no-sync-email", false, "E-Mail aus CSV nicht nach ChurchTools übernehmen")
	inviteCmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Plan ohne Rückfrage ausführen (für Skripte)")
	inviteCmd.Flags().BoolVar(&ignoreNameMismatch, "ignore-name-mismatch", false, "Zeilen auch verarbeiten, wenn der Name in der CSV nicht zur Person in ChurchTools passt")
	inviteCmd.Flags().BoolVar(&allowSingleEmail, "allow-single-email", false, "Wenn ChurchTools keine Adressliste annimmt: primäre E-Mail ersetzen (bisherige Adresse geht verloren)")
}

func runInvite(cmd *cobra.Command) error {
	if err := requireNonEmptyFlag(cmd, "--csv (-f)", csvPath); err != nil {
		return err
	}
	if err := validatePathFlagValueForCmd(cmd, "--csv", csvPath); err != nil {
		return err
	}
	if err := config.ValidateDelayMS("--delay-ms", delayMS); err != nil {
		return usageErrorf(cmd, "%v", err)
	}

	cfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}

	entries, err := csvfile.Read(csvPath)
	if err != nil {
		return err
	}

	client, err := connectChurchTools(cfg)
	if err != nil {
		return err
	}
	// Leave groups joined only for this run, also when the run fails.
	defer leaveTemporaryGroups(client)

	if !skipPreJoin {
		if err := ensurePreJoinGroups(client, cfg); err != nil {
			return err
		}
	}

	if !skipPermissionRequest {
		if err := ensureInvitePermissions(client, !noSyncEmail); err != nil {
			return err
		}
	}

	delay := time.Duration(cfg.EffectiveDelayMS()) * time.Millisecond
	if delayMS > 0 {
		delay = time.Duration(delayMS) * time.Millisecond
	}

	opts := invite.Options{
		DryRun:             dryRun,
		Delay:              delay,
		SyncEmail:          !noSyncEmail,
		Reinvite:           reinvite,
		AllowSingleEmail:   allowSingleEmail,
		IgnoreNameMismatch: ignoreNameMismatch,
		Context:            runCtx,
		Report:             func(result invite.Result) { invite.PrintResult(os.Stdout, result) },
	}

	if client.UsesMainInstance() {
		fmt.Fprintf(os.Stderr, "Warnung: API-Aufrufe laufen über die Hauptinstanz %s statt %s – die Personen-IDs der CSV müssen von dort stammen.\n", client.BaseURL(), cfg.BaseURL)
	}

	if dryRun {
		fmt.Printf("Dry-Run für %d Personen …\n", len(entries))
	} else {
		proceed, err := confirmInvitePlan(client, entries, opts)
		if err != nil {
			return err
		}
		if !proceed {
			fmt.Println("Abgebrochen – es wurde nichts gesendet oder geändert.")
			return nil
		}
		fmt.Printf("Sende Einladungen an %d Personen …\n", len(entries))
	}

	results, runErr := invite.Run(client, entries, opts)
	invite.PrintSummary(os.Stdout, results, len(entries))
	if runErr != nil {
		if errors.Is(runErr, context.Canceled) {
			return errInterrupted
		}
		return runErr
	}

	for _, result := range results {
		if !result.Success {
			return fmt.Errorf("mindestens ein Datensatz ist fehlgeschlagen")
		}
	}
	if !dryRun {
		printDeletionHint(csvPath)
	}
	return nil
}
