package churchtools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func noSleep(t *testing.T) {
	t.Helper()
	orig := sleep
	sleep = func(time.Duration) {}
	t.Cleanup(func() { sleep = orig })
}

func writeWhoAmI(w http.ResponseWriter) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]any{"id": 1, "firstName": "A", "lastName": "B", "email": "a@b.de"},
	})
}

func TestCallAPIRetriesOnceAfter429(t *testing.T) {
	noSleep(t)
	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) { writeWhoAmI(w) })
	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf"})
	})
	mux.HandleFunc("/api/persons/7", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": 7, "firstName": "X"}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "token", "", "")
	if err := client.Login(); err != nil {
		t.Fatal(err)
	}
	person, err := client.PersonByID(7)
	if err != nil || person.ID != 7 {
		t.Fatalf("PersonByID = %+v, %v", person, err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

// A password session that expires must be restored with username/password.
func TestPasswordSessionReloginAfter401(t *testing.T) {
	logins := 0
	sessionValid := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		logins++
		sessionValid = true
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"status": "success"}})
	})
	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		if !sessionValid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		writeWhoAmI(w)
	})
	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf"})
	})
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		if !sessionValid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "", "user", " pass word ")
	if err := client.Login(); err != nil {
		t.Fatal(err)
	}
	sessionValid = false // session expires

	if _, err := client.getAPI("/test", nil); err != nil {
		t.Fatalf("getAPI after expiry: %v", err)
	}
	if logins != 2 {
		t.Fatalf("logins = %d, want 2 (initial + relogin)", logins)
	}
}

func TestNewAPIErrorTruncatesBody(t *testing.T) {
	err := newAPIError(500, "", []byte(strings.Repeat("x", 10_000)))
	if len(err.Body) > 2100 {
		t.Fatalf("body length = %d, want ≤ ~2 KB", len(err.Body))
	}
}

// Password sessions must send the CSRF token on write requests (REST and
// legacy AJAX) under the header name "CSRF-Token" – not on GET or /api/login.
func TestCSRFHeaderOnWriteRequests(t *testing.T) {
	seen := map[string]string{}
	record := func(w http.ResponseWriter, r *http.Request) {
		seen[r.Method+" "+r.URL.Path] = r.Header.Get("CSRF-Token")
		if r.Header.Get("X-CSRF-Token") != "" {
			t.Errorf("%s %s: X-CSRF-Token must not be used", r.Method, r.URL.Path)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		record(w, r)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"status": "success"}})
	})
	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		record(w, r)
		writeWhoAmI(w)
	})
	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf-secret"})
	})
	mux.HandleFunc("/api/persons/10/invite", func(w http.ResponseWriter, r *http.Request) {
		record(w, r)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/persons/10", func(w http.ResponseWriter, r *http.Request) {
		record(w, r)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": 10}})
	})
	mux.HandleFunc("/api/groups/3/members/10", func(w http.ResponseWriter, r *http.Request) {
		record(w, r)
		if r.Method == http.MethodGet {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "", "user", "pass")
	if err := client.Login(); err != nil {
		t.Fatal(err)
	}
	if err := client.InvitePerson(10); err != nil {
		t.Fatalf("InvitePerson: %v", err)
	}
	if _, err := client.UpdatePersonEmail(10, EmailSyncPlan{Needed: true, Primary: "neu@example.org"}, false); err != nil {
		t.Fatalf("UpdatePersonEmail: %v", err)
	}
	if _, err := client.RequestGroupMembership(3, 10); err != nil {
		t.Fatalf("RequestGroupMembership: %v", err)
	}

	want := map[string]string{
		"POST /api/login":              "",
		"GET /api/whoami":              "",
		"GET /api/groups/3/members/10": "",
		"POST /api/persons/10/invite":  "csrf-secret",
		"PATCH /api/persons/10":        "csrf-secret",
		"PUT /api/groups/3/members/10": "csrf-secret",
	}
	for key, header := range want {
		got, ok := seen[key]
		if !ok {
			t.Errorf("%s: request not seen", key)
			continue
		}
		if got != header {
			t.Errorf("%s: CSRF-Token = %q, want %q", key, got, header)
		}
	}
}
