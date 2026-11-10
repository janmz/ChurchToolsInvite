package churchtools_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

// CR-5: without a primary address the other "emails" entries are kept,
// because the PATCH replaces the whole list.
func TestPrepareEmailUpdateWithoutPrimaryKeepsOtherAddresses(t *testing.T) {
	person := churchtools.Person{
		Email: "",
		Emails: []churchtools.PersonEmail{
			{Email: "dienst@example.org", IsDefault: false, ContactLabelID: 3},
		},
	}
	plan := churchtools.PrepareEmailUpdate("neu@example.org", person)
	if !plan.Needed || plan.Primary != "neu@example.org" {
		t.Fatalf("plan = %+v", plan)
	}
	var kept, newDefault bool
	for _, e := range plan.Emails {
		if e.Email == "dienst@example.org" && !e.IsDefault {
			kept = true
		}
		if e.Email == "neu@example.org" && e.IsDefault {
			newDefault = true
		}
	}
	if !kept || !newDefault {
		t.Fatalf("emails = %+v", plan.Emails)
	}
}

// CR-9: the single-field fallback is only used for 400/422 and reported.
func TestUpdatePersonEmailFallbackOnlyOnInvalidRequest(t *testing.T) {
	for _, tc := range []struct {
		status     int
		wantLegacy bool
		wantErr    bool
	}{
		{http.StatusUnprocessableEntity, true, false},
		{http.StatusInternalServerError, false, true},
	} {
		patches := 0
		mux := http.NewServeMux()
		mux.HandleFunc("/api/persons/7", func(w http.ResponseWriter, r *http.Request) {
			patches++
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if _, hasList := body["emails"]; hasList {
				http.Error(w, `{"message":"emails invalid"}`, tc.status)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		client := newLoggedInClient(t, mux)

		plan := churchtools.PrepareEmailUpdate("neu@example.org", churchtools.Person{
			Email:  "alt@example.org",
			Emails: []churchtools.PersonEmail{{Email: "alt@example.org", IsDefault: true}},
		})
		// S5: without permission the fallback is refused.
		if tc.wantLegacy {
			if _, err := client.UpdatePersonEmail(7, plan, false); !errors.Is(err, churchtools.ErrSingleEmailOnly) {
				t.Fatalf("status %d without allowSingle: err = %v", tc.status, err)
			}
			patches = 0
		}
		legacy, err := client.UpdatePersonEmail(7, plan, true)
		if legacy != tc.wantLegacy || (err != nil) != tc.wantErr {
			t.Fatalf("status %d: legacy = %v, err = %v", tc.status, legacy, err)
		}
		if !tc.wantLegacy && patches != 1 {
			t.Fatalf("status %d: %d PATCH requests, want 1 (no fallback)", tc.status, patches)
		}
	}
}

// CR-12: the server's reason is part of the error.
func TestInviteErrorContainsServerMessage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/persons/8/invite", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Person ist bereits Benutzer"}`))
	})
	client := newLoggedInClient(t, mux)

	err := client.InvitePerson(8)
	if err == nil || !strings.Contains(err.Error(), "Person ist bereits Benutzer") {
		t.Fatalf("err = %v", err)
	}
}

// CR-13: ambiguous group names are reported instead of "not found".
func TestFindGroupByNamesReportsAmbiguity(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/groups", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{
			{"id": 1, "name": "Personen bearbeiten"},
			{"id": 2, "name": "Personen bearbeiten"},
		}})
	})
	client := newLoggedInClient(t, mux)

	_, _, err := client.FindGroupByNames([]string{"Personen bearbeiten"})
	if err == nil || errors.Is(err, churchtools.ErrGroupNotFound) || !strings.Contains(err.Error(), "mehrere") {
		t.Fatalf("err = %v", err)
	}
}
