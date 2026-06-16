package invite_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	if results[0].Message == "" || results[0].Message[:8] != "dry-run:" {
		t.Fatalf("expected dry-run skip message, got %q", results[0].Message)
	}
}
