package cmd

import (
	"fmt"
	"time"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
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
	reinvite              bool
)

var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Einladungen für alle Personen in der CSV senden",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(runInvite())
	},
}

func init() {
	inviteCmd.Flags().StringVarP(&csvPath, "csv", "f", "", "Pfad zur CSV-Datei (Pflicht)")
	inviteCmd.Flags().IntVar(&delayMS, "delay-ms", 0, "Pause zwischen Einladungen in Millisekunden (0 = config.delay_ms)")
	inviteCmd.Flags().BoolVar(&skipPermissionRequest, "skip-permission-request", false, "Keine Gruppenmitgliedschaft für fehlende Berechtigungen beantragen")
	inviteCmd.Flags().BoolVar(&reinvite, "reinvite", false, "Bereits eingeladene Personen erneut einladen (Standard: überspringen)")
	inviteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Prüfen/simulieren ohne Einladungen zu senden oder ChurchTools zu ändern")
	inviteCmd.Flags().BoolVar(&noSyncEmail, "no-sync-email", false, "E-Mail aus CSV nicht nach ChurchTools übernehmen")
	_ = inviteCmd.MarkFlagRequired("csv")
}

func runInvite() error {
	if csvPath == "" {
		return fmt.Errorf("--csv ist erforderlich")
	}

	cfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}

	entries, err := csvfile.Read(csvPath)
	if err != nil {
		return err
	}

	client := churchtools.NewClient(cfg.BaseURL, cfg.LoginToken, cfg.Username, cfg.Password)
	if err := client.Login(); err != nil {
		return err
	}

	if !skipPermissionRequest {
		if err := ensureInvitePermissions(client, cfg, !noSyncEmail); err != nil {
			return err
		}
	}

	delay := time.Duration(cfg.DelayMS) * time.Millisecond
	if delayMS > 0 {
		delay = time.Duration(delayMS) * time.Millisecond
	}

	opts := invite.Options{
		DryRun:    dryRun,
		Delay:     delay,
		SyncEmail: !noSyncEmail,
		Reinvite:  reinvite,
	}

	if dryRun {
		fmt.Printf("Dry-Run für %d Personen …\n", len(entries))
	} else {
		fmt.Printf("Sende Einladungen an %d Personen …\n", len(entries))
	}

	results, err := invite.Run(client, entries, opts)
	if err != nil {
		return err
	}

	invite.PrintSummary(results)

	for _, result := range results {
		if !result.Success {
			return fmt.Errorf("mindestens ein datensatz ist fehlgeschlagen")
		}
	}
	return nil
}
