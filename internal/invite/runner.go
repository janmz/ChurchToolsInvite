package invite

import (
	"fmt"
	"strings"
	"time"

	"github.com/janmz/masseneinladung/internal/churchtools"
	"github.com/janmz/masseneinladung/internal/csvfile"
)

// Result describes the outcome for one CSV row.
type Result struct {
	Entry   csvfile.Entry
	Success bool
	Message string
}

// Options controls batch invitation behaviour.
type Options struct {
	DryRun       bool
	Delay        time.Duration
	ValidateOnly bool
	SyncEmail    bool
}

// Run sends invitations for all CSV entries.
func Run(client *churchtools.Client, entries []csvfile.Entry, opts Options) ([]Result, error) {
	results := make([]Result, 0, len(entries))

	for i, entry := range entries {
		if i > 0 && opts.Delay > 0 {
			time.Sleep(opts.Delay)
		}

		result := Result{Entry: entry}
		label := formatLabel(entry)

		person, err := client.PersonByID(entry.PersonID)
		if err != nil {
			result.Message = fmt.Sprintf("person laden fehlgeschlagen: %v", err)
			results = append(results, result)
			continue
		}

		inviteEmail, emailNote, err := resolveInviteEmail(client, person, entry, opts)
		if err != nil {
			result.Message = err.Error()
			results = append(results, result)
			continue
		}

		if opts.ValidateOnly || opts.DryRun {
			if inviteEmail == "" {
				result.Message = "person hat keine e-mail-adresse (weder csv noch churchtools)"
				results = append(results, result)
				continue
			}

			if opts.DryRun {
				result.Success = true
				result.Message = fmt.Sprintf("dry-run: würde %s (%s) einladen%s", label, inviteEmail, emailNote)
				results = append(results, result)
				continue
			}

			result.Success = true
			result.Message = fmt.Sprintf("ok: %s (%s)%s", label, inviteEmail, emailNote)
			results = append(results, result)
			continue
		}

		if inviteEmail == "" {
			result.Message = "person hat keine e-mail-adresse (weder csv noch churchtools)"
			results = append(results, result)
			continue
		}

		if err := client.InvitePerson(entry.PersonID); err != nil {
			result.Message = err.Error()
			results = append(results, result)
			continue
		}

		result.Success = true
		result.Message = fmt.Sprintf("einladung gesendet an %s (%s)%s", label, inviteEmail, emailNote)
		results = append(results, result)
	}

	return results, nil
}

func resolveInviteEmail(
	client *churchtools.Client,
	person churchtools.Person,
	entry csvfile.Entry,
	opts Options,
) (string, string, error) {
	csvEmail := strings.TrimSpace(entry.Email)
	inviteEmail := strings.TrimSpace(person.Email)
	note := ""

	if csvEmail == "" {
		return inviteEmail, note, nil
	}

	if !opts.SyncEmail {
		if inviteEmail != "" && !strings.EqualFold(csvEmail, inviteEmail) {
			return "", "", fmt.Errorf("e-mail weicht ab (csv=%s, ct=%s)", csvEmail, inviteEmail)
		}
		if inviteEmail == "" {
			inviteEmail = csvEmail
		}
		return inviteEmail, note, nil
	}

	plan := churchtools.PrepareEmailUpdate(csvEmail, person)
	if !plan.Needed {
		if inviteEmail == "" {
			inviteEmail = csvEmail
		}
		return inviteEmail, note, nil
	}

	if opts.DryRun || opts.ValidateOnly {
		note = "; " + plan.Detail
		return plan.Primary, note, nil
	}

	if err := client.UpdatePersonEmail(entry.PersonID, plan); err != nil {
		return "", "", fmt.Errorf("e-mail-sync fehlgeschlagen: %w", err)
	}

	note = "; " + plan.Detail
	return plan.Primary, note, nil
}

func formatLabel(entry csvfile.Entry) string {
	name := strings.TrimSpace(strings.Join([]string{entry.FirstName, entry.LastName}, " "))
	if name == "" {
		name = "Person"
	}
	return fmt.Sprintf("%s (ID %d)", name, entry.PersonID)
}

// PrintSummary writes a human-readable report to stdout via fmt.
func PrintSummary(results []Result) {
	success := 0
	for _, result := range results {
		if result.Success {
			success++
		}
	}

	fmt.Printf("\nZusammenfassung: %d/%d erfolgreich\n", success, len(results))
	for _, result := range results {
		status := "FEHLER"
		if result.Success {
			status = "OK"
		}
		fmt.Printf("[%s] Zeile %d, ID %d: %s\n",
			status,
			result.Entry.Line,
			result.Entry.PersonID,
			result.Message,
		)
	}
}
