package invite_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	csvfile "github.com/janmz/churchtools-invite/internal/csvfile"
	invite "github.com/janmz/churchtools-invite/internal/invite"
)

func TestDryRunSkipsAlreadyInvited(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": 1, "firstName": "Admin", "lastName": "User", "email": "admin@example.org"},
		})
	})
	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf-test"})
	})
	mux.HandleFunc("/api/persons/99", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id":               99,
				"firstName":        "Max",
				"lastName":         "Muster",
				"email":            "max@example.org",
				"invitationStatus": "accepted",
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := churchtools.NewClient(server.URL, "test-token", "", "")
	if err := client.Login(); err != nil {
		t.Fatalf("Login: %v", err)
	}

	results, err := invite.Run(client, []csvfile.Entry{{PersonID: 99, Line: 2}}, invite.Options{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !results[0].Skipped {
		t.Fatalf("expected skipped result, got %+v", results)
	}
	if results[0].Message == "" || !strings.HasPrefix(results[0].Message, "dry-run:") {
		t.Fatalf("expected dry-run skip message, got %q", results[0].Message)
	}
}

func TestReinviteSkipsRegisteredUser(t *testing.T) {
	invited := false
	mux := http.NewServeMux()
	registerInviteMocks(mux, map[string]any{
		"id":               99,
		"firstName":        "Max",
		"lastName":         "Muster",
		"email":            "max@example.org",
		"invitationStatus": "accepted",
	})
	mux.HandleFunc("/api/persons/99/invite", func(w http.ResponseWriter, r *http.Request) {
		invited = true
		w.WriteHeader(http.StatusNoContent)
	})

	client := newInviteTestClient(t, mux)
	results, err := invite.Run(client, []csvfile.Entry{{
		Line: 2, PersonID: 99, Email: "max@example.org",
	}}, invite.Options{Reinvite: true})
	if err != nil {
		t.Fatal(err)
	}
	if invited {
		t.Fatal("registered user must not be invited even with --reinvite")
	}
	if len(results) != 1 || !results[0].Skipped {
		t.Fatalf("expected skipped result, got %+v", results)
	}
}

func TestReinvitePendingUser(t *testing.T) {
	invited := false
	mux := http.NewServeMux()
	registerInviteMocks(mux, map[string]any{
		"id":               99,
		"firstName":        "Max",
		"lastName":         "Muster",
		"email":            "max@example.org",
		"invitationStatus": "pending",
	})
	mux.HandleFunc("/api/persons/99/invite", func(w http.ResponseWriter, r *http.Request) {
		invited = true
		w.WriteHeader(http.StatusNoContent)
	})

	client := newInviteTestClient(t, mux)
	results, err := invite.Run(client, []csvfile.Entry{{
		Line: 2, PersonID: 99, Email: "max@example.org",
	}}, invite.Options{Reinvite: true})
	if err != nil {
		t.Fatal(err)
	}
	if !invited {
		t.Fatal("pending user should be re-invited with --reinvite")
	}
	if len(results) != 1 || !results[0].Success || results[0].Skipped {
		t.Fatalf("unexpected result: %+v", results[0])
	}
}

func TestDryRunReinvitesWhenEmailDiffers(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": 1, "firstName": "Admin", "lastName": "User", "email": "admin@example.org"},
		})
	})
	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf-test"})
	})
	mux.HandleFunc("/api/persons/99", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id":               99,
				"firstName":        "Max",
				"lastName":         "Muster",
				"email":            "alt@example.org",
				"invitationStatus": "pending",
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := churchtools.NewClient(server.URL, "test-token", "", "")
	if err := client.Login(); err != nil {
		t.Fatalf("Login: %v", err)
	}

	entry := csvfile.Entry{Line: 2, PersonID: 99, Email: "neu@example.org"}
	results, err := invite.Run(client, []csvfile.Entry{entry}, invite.Options{DryRun: true, SyncEmail: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Skipped {
		t.Fatalf("expected invite result, got %+v", results)
	}
	if !strings.Contains(results[0].Message, "neu@example.org") {
		t.Fatalf("expected new email in message, got %q", results[0].Message)
	}
}

func TestLiveInviteSuccess(t *testing.T) {
	invited := false
	mux := http.NewServeMux()
	registerInviteMocks(mux, map[string]any{
		"id":        99,
		"firstName": "Max",
		"lastName":  "Muster",
		"email":     "max@example.org",
	})
	mux.HandleFunc("/api/persons/99/invite", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		invited = true
		w.WriteHeader(http.StatusNoContent)
	})

	client := newInviteTestClient(t, mux)
	results, err := invite.Run(client, []csvfile.Entry{{
		Line: 2, PersonID: 99, Email: "max@example.org",
	}}, invite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !invited {
		t.Fatal("expected invite API call")
	}
	if len(results) != 1 || !results[0].Success || results[0].Skipped {
		t.Fatalf("unexpected result: %+v", results[0])
	}
}

func TestEmailMismatchWithoutSync(t *testing.T) {
	mux := http.NewServeMux()
	registerInviteMocks(mux, map[string]any{
		"id":    99,
		"email": "alt@example.org",
	})

	client := newInviteTestClient(t, mux)
	results, err := invite.Run(client, []csvfile.Entry{{
		Line: 2, PersonID: 99, Email: "neu@example.org",
	}}, invite.Options{SyncEmail: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Success {
		t.Fatalf("expected failure, got %+v", results[0])
	}
	if !strings.Contains(results[0].Message, "e-mail weicht ab") {
		t.Fatalf("unexpected message: %q", results[0].Message)
	}
}

func TestSyncEmailForbiddenStillInvites(t *testing.T) {
	invited := false
	mux := http.NewServeMux()
	registerAuthMocks(mux)
	mux.HandleFunc("/api/persons/99", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{
				"id":    99,
				"email": "alt@example.org",
				"emails": []map[string]any{{
					"email":     "alt@example.org",
					"isDefault": true,
				}},
			}})
		case http.MethodPatch:
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.Error(w, "method", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/persons/99/invite", func(w http.ResponseWriter, r *http.Request) {
		invited = true
		w.WriteHeader(http.StatusNoContent)
	})

	client := newInviteTestClient(t, mux)
	results, err := invite.Run(client, []csvfile.Entry{{
		Line: 2, PersonID: 99, Email: "neu@example.org",
	}}, invite.Options{SyncEmail: true})
	if err != nil {
		t.Fatal(err)
	}
	if !invited {
		t.Fatal("expected invite despite forbidden email sync")
	}
	if len(results) != 1 || !results[0].Success {
		t.Fatalf("unexpected result: %+v", results[0])
	}
	if !strings.Contains(results[0].Message, "e-mail-sync übersprungen") {
		t.Fatalf("expected fallback note, got %q", results[0].Message)
	}
}

func registerInviteMocks(mux *http.ServeMux, person map[string]any) {
	registerAuthMocks(mux)
	mux.HandleFunc("/api/persons/"+formatPersonID(person["id"]), func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": person})
	})
}

func registerAuthMocks(mux *http.ServeMux) {
	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": 1, "firstName": "Admin", "lastName": "User", "email": "admin@example.org"},
		})
	})
	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf-test"})
	})
}

func newInviteTestClient(t *testing.T, mux *http.ServeMux) *churchtools.Client {
	t.Helper()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client := churchtools.NewClient(server.URL, "test-token", "", "")
	if err := client.Login(); err != nil {
		t.Fatalf("Login: %v", err)
	}
	return client
}

func formatPersonID(id any) string {
	return fmt.Sprint(id)
}
