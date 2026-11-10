package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	csvfile "github.com/janmz/churchtools-invite/internal/csvfile"
	invite "github.com/janmz/churchtools-invite/internal/invite"
	"github.com/janmz/churchtools-invite/internal/textnorm"
)

// invitePlan summarises a dry-run of the CSV before the real run.
type invitePlan struct {
	invitations  []invite.Result
	emailChanges []invite.Result
	skipped      int
	failed       []invite.Result
}

func buildInvitePlan(results []invite.Result) invitePlan {
	var plan invitePlan
	for _, result := range results {
		switch {
		case result.WouldInvite:
			plan.invitations = append(plan.invitations, result)
			if result.EmailChange != "" {
				plan.emailChanges = append(plan.emailChanges, result)
			}
		case result.Skipped:
			plan.skipped++
		case !result.Success:
			plan.failed = append(plan.failed, result)
		}
	}
	return plan
}

// print shows the plan with full addresses, so the operator can check that
// every invitation goes to the right person.
func (p invitePlan) print(w io.Writer) {
	fmt.Fprintf(w, "\nPlan: %d Einladungen, davon %d mit E-Mail-Änderung; %d übersprungen, %d fehlerhaft (werden nicht verarbeitet).\n",
		len(p.invitations), len(p.emailChanges), p.skipped, len(p.failed))
	if len(p.emailChanges) > 0 {
		fmt.Fprintln(w, "E-Mail-Änderungen (die Einladung geht an die neue Adresse):")
		for _, result := range p.emailChanges {
			fmt.Fprintf(w, "  Zeile %d, ID %d, %s %s: %s\n", result.Entry.Line, result.Entry.PersonID,
				textnorm.Printable(result.Entry.FirstName), textnorm.Printable(result.Entry.LastName),
				textnorm.Printable(result.EmailChange))
		}
	}
	if len(p.failed) > 0 {
		fmt.Fprintln(w, "Fehlerhafte Zeilen:")
		for _, result := range p.failed {
			invite.PrintResult(w, result)
		}
	}
}

// confirmInvitePlan runs the CSV as a dry-run, shows what would be sent and
// changed and asks for confirmation. --yes skips the question, but not when
// the API calls were moved to the main instance (IDs may refer to other
// persons there).
func confirmInvitePlan(client *churchtools.Client, entries []csvfile.Entry, opts invite.Options) (bool, error) {
	planOpts := opts
	planOpts.DryRun = true
	planOpts.Report = nil
	planOpts.Delay = 0

	fmt.Fprintf(os.Stderr, "Prüfe %d Zeilen (Plan, noch keine Änderung) …\n", len(entries))
	results, err := invite.Run(client, entries, planOpts)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return false, errInterrupted
		}
		return false, err
	}

	plan := buildInvitePlan(results)
	plan.print(os.Stderr)
	if len(plan.invitations) == 0 {
		fmt.Fprintln(os.Stderr, "Nichts zu senden.")
		return false, nil
	}

	if assumeYes {
		if client.UsesMainInstance() {
			return false, fmt.Errorf("Lauf über die Hauptinstanz %s: --yes ist hier nicht erlaubt, bitte interaktiv bestätigen", client.BaseURL())
		}
		return true, nil
	}

	answer, err := readLine(fmt.Sprintf("%d Einladungen jetzt senden? [j/N]: ", len(plan.invitations)))
	if err != nil {
		if errors.Is(err, io.EOF) {
			return false, fmt.Errorf("keine Bestätigung möglich (keine Eingabe) – Plan prüfen und mit --yes ausführen")
		}
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(answer), "j"), nil
}
