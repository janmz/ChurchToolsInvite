package invite_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	csvfile "github.com/janmz/churchtools-invite/internal/csvfile"
	invite "github.com/janmz/churchtools-invite/internal/invite"
)

// CR-6: with --no-sync-email ChurchTools invites to its own address; without
// one the row fails already in the dry-run.
func TestNoSyncWithoutChurchToolsEmailFails(t *testing.T) {
	mux := http.NewServeMux()
	registerInviteMocks(mux, map[string]any{"id": 99, "firstName": "Max", "lastName": "Muster", "email": ""})
	client := newInviteTestClient(t, mux)

	results, err := invite.Run(client, []csvfile.Entry{{Line: 2, PersonID: 99, Email: "max@example.org"}},
		invite.Options{DryRun: true, SyncEmail: false})
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Success || !strings.Contains(results[0].Message, "keine E-Mail-Adresse in ChurchTools") {
		t.Fatalf("result = %+v", results[0])
	}
}

// CR-15: a real run reports "übersprungen", not "würde überspringen".
func TestRealRunSkipMessage(t *testing.T) {
	mux := http.NewServeMux()
	registerInviteMocks(mux, map[string]any{
		"id": 99, "email": "max@example.org", "invitationStatus": "accepted",
	})
	client := newInviteTestClient(t, mux)

	results, err := invite.Run(client, []csvfile.Entry{{Line: 2, PersonID: 99}}, invite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(results[0].Message, "übersprungen:") {
		t.Fatalf("message = %q", results[0].Message)
	}
}

// CR-19, CR-20, CR-8: persons are loaded with one list request, the delay is
// only waited after an invitation and every result is reported at once.
func TestRunPreloadsPersonsAndDelaysOnlyAfterWrites(t *testing.T) {
	var sleeps int
	defer invite.SetSleepForTest(func(time.Duration) { sleeps++ })()

	detailCalls := 0
	invites := 0
	mux := http.NewServeMux()
	registerAuthMocks(mux)
	mux.HandleFunc("/api/persons", func(w http.ResponseWriter, r *http.Request) {
		if got := len(r.URL.Query()["ids[]"]); got != 3 {
			t.Errorf("ids[] count = %d, want 3", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{
			{"id": 1, "email": "a@example.org", "invitationStatus": "accepted"},
			{"id": 2, "email": "b@example.org"},
			{"id": 3, "email": "c@example.org", "invitationStatus": "accepted"},
		}})
	})
	mux.HandleFunc("/api/persons/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/invite") {
			invites++
			w.WriteHeader(http.StatusNoContent)
			return
		}
		detailCalls++
		http.NotFound(w, r)
	})
	client := newInviteTestClient(t, mux)

	var reported []int
	entries := []csvfile.Entry{{Line: 2, PersonID: 1}, {Line: 3, PersonID: 2}, {Line: 4, PersonID: 3}}
	results, err := invite.Run(client, entries, invite.Options{
		Delay:  time.Second,
		Report: func(r invite.Result) { reported = append(reported, r.Entry.PersonID) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if detailCalls != 0 {
		t.Fatalf("%d single person requests, want 0", detailCalls)
	}
	if invites != 1 || !results[1].Success || !results[0].Skipped || !results[2].Skipped {
		t.Fatalf("invites = %d, results = %+v", invites, results)
	}
	if sleeps != 1 {
		t.Fatalf("sleeps = %d, want 1 (only after the invitation)", sleeps)
	}
	if len(reported) != 3 {
		t.Fatalf("reported = %v", reported)
	}
}

// CR-4: a cancelled context stops the run before the next entry.
func TestRunStopsWhenCancelled(t *testing.T) {
	mux := http.NewServeMux()
	registerInviteMocks(mux, map[string]any{"id": 99, "email": "max@example.org", "invitationStatus": "accepted"})
	client := newInviteTestClient(t, mux)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results, err := invite.Run(client, []csvfile.Entry{{Line: 2, PersonID: 99}}, invite.Options{Context: ctx})
	if err == nil || len(results) != 0 {
		t.Fatalf("results = %+v, err = %v", results, err)
	}
}
