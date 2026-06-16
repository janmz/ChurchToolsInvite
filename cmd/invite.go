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
	csvPath     string
	dryRun      bool
	delayMS     int
	noSyncEmail bool
)

var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Einladungen für alle Personen in der CSV senden",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(runInvite(false))
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "CSV prüfen ohne Einladungen zu senden",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(runInvite(true))
	},
}

func init() {
	for _, command := range []*cobra.Command{inviteCmd, validateCmd} {
		command.Flags().StringVarP(&csvPath, "csv", "f", "", "Pfad zur CSV-Datei (Pflicht)")
		command.Flags().IntVar(&delayMS, "delay-ms", 0, "Pause zwischen Einladungen in Millisekunden (0 = config.delay_ms)")
		_ = command.MarkFlagRequired("csv")
	}
	inviteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Nur simulieren, keine E-Mails senden")
	inviteCmd.Flags().BoolVar(&noSyncEmail, "no-sync-email", false, "E-Mail aus CSV nicht nach ChurchTools übernehmen")
}

func runInvite(validateOnly bool) error {
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

	delay := time.Duration(cfg.DelayMS) * time.Millisecond
	if delayMS > 0 {
		delay = time.Duration(delayMS) * time.Millisecond
	}

	opts := invite.Options{
		DryRun:       dryRun,
		Delay:        delay,
		ValidateOnly: validateOnly,
		SyncEmail:    !noSyncEmail,
	}

	if validateOnly {
		fmt.Printf("Validiere %d Personen …\n", len(entries))
	} else if dryRun {
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
