package churchtools_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

func TestLoginAndInvite(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Login test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id":        1,
				"campusId":  3,
				"firstName": "Admin",
				"lastName":  "User",
				"email":     "admin@example.org",
			},
		})
	})

	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf-test"})
	})

	mux.HandleFunc("/api/persons/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":        99,
					"firstName": "Max",
					"lastName":  "Muster",
					"email":     "alt@example.org",
					"emails": []map[string]any{{
						"email":          "alt@example.org",
						"isDefault":      true,
						"contactLabelId": 2,
					}},
				},
			})
		case r.Method == http.MethodPatch:
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "bad json", http.StatusBadRequest)
				return
			}
			if payload["email"] != "neu@example.org" {
				http.Error(w, "unexpected email", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/api/persons/99/invite":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method", http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := churchtools.NewClient(server.URL, "test-token", "", "")
	if err := client.Login(); err != nil {
		t.Fatalf("Login: %v", err)
	}

	campusID, err := client.CurrentUserCampusID()
	if err != nil {
		t.Fatalf("CurrentUserCampusID: %v", err)
	}
	if campusID != 3 {
		t.Fatalf("campusID = %d", campusID)
	}

	if err := client.InvitePerson(99); err != nil {
		t.Fatalf("InvitePerson: %v", err)
	}
}

func TestUpdatePersonEmail(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": 1, "firstName": "Admin", "lastName": "User", "email": "admin@example.org"},
		})
	})

	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf-test"})
	})

	mux.HandleFunc("/api/persons/42", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "json", http.StatusBadRequest)
			return
		}
		if payload["email"] != "neu@example.org" {
			http.Error(w, "email", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := churchtools.NewClient(server.URL, "test-token", "", "")
	if err := client.Login(); err != nil {
		t.Fatalf("Login: %v", err)
	}

	plan := churchtools.EmailSyncPlan{
		Needed:  true,
		Primary: "neu@example.org",
		Emails: []churchtools.PersonEmail{
			{Email: "neu@example.org", IsDefault: true, ContactLabelID: 2},
			{Email: "alt@example.org", IsDefault: false, ContactLabelID: 2},
		},
	}
	if err := client.UpdatePersonEmail(42, plan); err != nil {
		t.Fatalf("UpdatePersonEmail: %v", err)
	}
}

func TestFindInvitePermissions(t *testing.T) {
	perms := map[string]any{
		"administration": map[string]any{
			"invitePersons": true,
		},
		"notes": []any{"Gruppenmitglieder zu ChurchTools einladen"},
	}

	found := churchtools.FindInvitePermissions(perms)
	if len(found) < 2 {
		t.Fatalf("expected matches, got %v", found)
	}
}

func TestIsForbidden(t *testing.T) {
	if !churchtools.IsForbidden(&churchtools.APIError{StatusCode: 403}) {
		t.Fatal("expected forbidden")
	}
	if churchtools.IsForbidden(&churchtools.APIError{StatusCode: 404}) {
		t.Fatal("expected not forbidden")
	}
}
