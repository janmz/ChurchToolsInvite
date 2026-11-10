package invite

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	csvfile "github.com/janmz/churchtools-invite/internal/csvfile"
	"github.com/janmz/churchtools-invite/internal/textnorm"
)

// Result describes the outcome for one CSV row.
type Result struct {
	Entry   csvfile.Entry
	Success bool
	Skipped bool
	Message string
	// WouldInvite is set in a dry-run for rows that would get an invitation.
	WouldInvite bool
	// EmailChange describes a (planned) e-mail update ("alt → neu").
	EmailChange string
}

// Options controls batch invitation behaviour.
type Options struct {
	DryRun    bool
	Delay     time.Duration
	SyncEmail bool
	Reinvite  bool
	// AllowSingleEmail permits replacing the primary address when the
	// instance does not accept the "emails" list (previous address is lost).
	AllowSingleEmail bool
	// IgnoreNameMismatch processes rows whose CSV name differs from the
	// ChurchTools person (normally they fail: shifted rows or typos in the ID
	// would otherwise send another person's account activation elsewhere).
	IgnoreNameMismatch bool
	// Context stops the run between two entries (e.g. Ctrl+C); nil = never.
	Context context.Context
	// Report receives each result as soon as it is known (progress output).
	Report func(Result)
}

// sleep waits for d or until ctx is done (replaceable in tests).
var sleep = func(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// Run sends invitations for all CSV entries. Persons are loaded in blocks
// up front; delay_ms is only waited after a write request (invitation or
// e-mail update), not after skipped or failed entries. When the context is
// cancelled, the results so far are returned together with ctx.Err().
func Run(client *churchtools.Client, entries []csvfile.Entry, opts Options) ([]Result, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	results := make([]Result, 0, len(entries))
	persons := preloadPersons(client, entries)
	wrote := false

	for _, entry := range entries {
		if wrote && opts.Delay > 0 {
			sleep(ctx, opts.Delay)
		}
		if err := ctx.Err(); err != nil {
			return results, err
		}

		result, didWrite := processEntry(client, persons, entry, opts)
		wrote = didWrite
		if !opts.DryRun {
			// Real runs may be logged; full addresses only in the dry-run.
			result.Message = MaskEmails(result.Message)
		}
		results = append(results, result)
		if opts.Report != nil {
			opts.Report(result)
		}
	}

	return results, nil
}

// preloadPersons loads all CSV persons with few list requests. On failure an
// empty map is returned and every person is loaded individually instead.
func preloadPersons(client *churchtools.Client, entries []csvfile.Entry) map[int]churchtools.Person {
	ids := make([]int, len(entries))
	for i, entry := range entries {
		ids[i] = entry.PersonID
	}
	byID := make(map[int]churchtools.Person, len(entries))
	persons, err := client.ListPersons(churchtools.PersonListOptions{IDs: ids})
	if err != nil {
		return byID
	}
	for _, person := range persons {
		byID[person.ID] = person
	}
	return byID
}

// processEntry handles one CSV row and reports whether ChurchTools was
// changed (invitation sent or e-mail updated).
func processEntry(
	client *churchtools.Client,
	persons map[int]churchtools.Person,
	entry csvfile.Entry,
	opts Options,
) (Result, bool) {
	result := Result{Entry: entry}
	label := formatLabel(entry)

	person, ok := persons[entry.PersonID]
	if !ok {
		var err error
		person, err = client.PersonByID(entry.PersonID)
		if err != nil {
			result.Message = fmt.Sprintf("Person laden fehlgeschlagen: %v", err)
			return result, false
		}
	}

	if !opts.IgnoreNameMismatch {
		if mismatch := nameMismatch(entry, person); mismatch != "" {
			result.Message = mismatch
			return result, false
		}
	}

	if shouldSkipInvitedPerson(person, entry, opts) {
		result.Success = true
		result.Skipped = true
		if opts.DryRun {
			result.Message = fmt.Sprintf("dry-run: würde überspringen: %s (%s)", label, person.AccountStatusLabel())
		} else {
			result.Message = fmt.Sprintf("übersprungen: %s (%s)", label, person.AccountStatusLabel())
		}
		return result, false
	}

	inviteEmail, emailNote, updated, err := resolveInviteEmail(client, person, entry, opts)
	if err != nil {
		result.Message = err.Error()
		return result, updated
	}
	if !strings.EqualFold(inviteEmail, strings.TrimSpace(person.Email)) && inviteEmail != "" {
		old := strings.TrimSpace(person.Email)
		if old == "" {
			old = "(keine)"
		}
		result.EmailChange = old + " → " + inviteEmail
	}
	if inviteEmail == "" {
		result.Message = "Person hat keine E-Mail-Adresse (weder CSV noch ChurchTools)"
		return result, updated
	}

	if opts.DryRun {
		result.Success = true
		result.WouldInvite = true
		result.Message = fmt.Sprintf("dry-run: würde %s (%s) einladen%s", label, inviteEmail, emailNote)
		return result, false
	}

	if err := client.InvitePerson(entry.PersonID); err != nil {
		result.Message = err.Error()
		return result, true
	}

	result.Success = true
	result.Message = fmt.Sprintf("Einladung gesendet an %s (%s)%s", label, inviteEmail, emailNote)
	return result, true
}

// resolveInviteEmail returns the address ChurchTools will send the
// invitation to (always the person's ChurchTools address, after an optional
// sync from the CSV), a note for the report and whether the person was
// changed.
func resolveInviteEmail(
	client *churchtools.Client,
	person churchtools.Person,
	entry csvfile.Entry,
	opts Options,
) (string, string, bool, error) {
	csvEmail := strings.TrimSpace(entry.Email)
	ctEmail := strings.TrimSpace(person.Email)

	if csvEmail == "" {
		return ctEmail, "", false, nil
	}

	if !opts.SyncEmail {
		if ctEmail == "" {
			return "", "", false, fmt.Errorf("Person hat keine E-Mail-Adresse in ChurchTools (CSV: %s; Abgleich mit --no-sync-email deaktiviert)", csvEmail)
		}
		if !strings.EqualFold(csvEmail, ctEmail) {
			return "", "", false, fmt.Errorf("e-mail weicht ab (csv=%s, ct=%s)", csvEmail, ctEmail)
		}
		return ctEmail, "", false, nil
	}

	plan := churchtools.PrepareEmailUpdate(csvEmail, person)
	if !plan.Needed {
		return ctEmail, "", false, nil
	}

	if opts.DryRun {
		return plan.Primary, "; " + plan.Detail, false, nil
	}

	legacy, err := client.UpdatePersonEmail(entry.PersonID, plan, opts.AllowSingleEmail)
	if err != nil {
		if churchtools.IsForbidden(err) {
			note := "; e-mail-sync übersprungen (keine Berechtigung Personen bearbeiten)"
			if ctEmail == "" {
				return "", "", false, fmt.Errorf("e-mail-sync nicht möglich (keine Berechtigung) und Person hat keine E-Mail-Adresse in ChurchTools")
			}
			if !strings.EqualFold(csvEmail, ctEmail) {
				note += fmt.Sprintf("; Einladung an ChurchTools-Adresse %s (CSV: %s)", ctEmail, csvEmail)
			}
			return ctEmail, note, false, nil
		}
		return "", "", false, fmt.Errorf("e-mail-sync fehlgeschlagen: %w", err)
	}

	if legacy {
		return plan.Primary, fmt.Sprintf("; primäre E-Mail auf %s gesetzt (nur Einzeladresse unterstützt, bisherige Adresse ersetzt)", plan.Primary), true, nil
	}
	return plan.Primary, "; " + plan.Detail, true, nil
}

func shouldSkipInvitedPerson(person churchtools.Person, entry csvfile.Entry, opts Options) bool {
	if person.IsRegisteredUser() {
		return true
	}
	if opts.Reinvite || !person.HasChurchToolsAccount() {
		return false
	}
	return !needsEmailUpdate(person, entry, opts)
}

func needsEmailUpdate(person churchtools.Person, entry csvfile.Entry, opts Options) bool {
	csvEmail := strings.TrimSpace(entry.Email)
	if csvEmail == "" {
		return false
	}
	if !opts.SyncEmail {
		current := strings.TrimSpace(person.Email)
		if current == "" {
			return true
		}
		return !strings.EqualFold(current, csvEmail)
	}
	return churchtools.PrepareEmailUpdate(csvEmail, person).Needed
}

// nameMismatch compares the CSV name with the ChurchTools person (umlauts,
// accents, case and punctuation ignored) and describes a difference. Empty
// CSV name fields are not compared.
func nameMismatch(entry csvfile.Entry, person churchtools.Person) string {
	first := strings.TrimSpace(entry.FirstName)
	last := strings.TrimSpace(entry.LastName)
	if first == "" && last == "" {
		return ""
	}
	firstOK := first == "" || textnorm.FoldAlnum(first) == textnorm.FoldAlnum(person.FirstName)
	lastOK := last == "" || textnorm.FoldAlnum(last) == textnorm.FoldAlnum(person.LastName)
	if firstOK && lastOK {
		return ""
	}
	return fmt.Sprintf("Name weicht ab (CSV: %s %s, ChurchTools: %s %s) – Zeile verrutscht oder ID falsch? Mit --ignore-name-mismatch trotzdem verarbeiten",
		first, last, person.FirstName, person.LastName)
}

// emailPattern finds e-mail addresses in report text.
var emailPattern = regexp.MustCompile(`[^\s@(),;:<>"']+@[^\s@(),;:<>"']+`)

// MaskEmails shortens the local part of all addresses in text
// ("max.muster@example.org" → "ma***@example.org").
func MaskEmails(text string) string {
	return emailPattern.ReplaceAllStringFunc(text, func(addr string) string {
		at := strings.LastIndex(addr, "@")
		local := []rune(addr[:at])
		keep := 2
		if len(local) < keep {
			keep = len(local)
		}
		return string(local[:keep]) + "***" + addr[at:]
	})
}

func formatLabel(entry csvfile.Entry) string {
	name := strings.TrimSpace(strings.Join([]string{entry.FirstName, entry.LastName}, " "))
	if name == "" {
		name = "Person"
	}
	return fmt.Sprintf("%s (ID %d)", name, entry.PersonID)
}

// PrintResult writes one result line (used as Options.Report).
func PrintResult(w io.Writer, result Result) {
	status := "FEHLER"
	switch {
	case result.Skipped:
		status = "ÜBERSPRUNGEN"
	case result.Success:
		status = "OK"
	}
	fmt.Fprintf(w, "[%s] Zeile %d, ID %d: %s\n",
		status,
		result.Entry.Line,
		result.Entry.PersonID,
		textnorm.Printable(result.Message),
	)
}

// PrintSummary writes the totals and repeats the failed rows, because each
// result was already printed while the run was in progress. total is the
// number of CSV rows (larger than len(results) after an abort).
func PrintSummary(w io.Writer, results []Result, total int) {
	success := 0
	skipped := 0
	var failed []Result
	for _, result := range results {
		if result.Success {
			success++
		} else {
			failed = append(failed, result)
		}
		if result.Skipped {
			skipped++
		}
	}

	fmt.Fprintf(w, "\nZusammenfassung: %d/%d erfolgreich", success, total)
	if skipped > 0 {
		fmt.Fprintf(w, " (%d übersprungen)", skipped)
	}
	if len(results) < total {
		fmt.Fprintf(w, ", abgebrochen nach %d von %d Zeilen", len(results), total)
	}
	fmt.Fprintln(w)

	if len(failed) > 0 {
		fmt.Fprintln(w, "Fehlgeschlagen:")
		for _, result := range failed {
			PrintResult(w, result)
		}
	}
}
