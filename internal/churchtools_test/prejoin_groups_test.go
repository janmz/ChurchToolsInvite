package churchtools_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

func TestEnsurePreJoinGroupsSkipsExistingMembership(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": 7, "firstName": "Admin", "lastName": "User", "email": "admin@example.org"},
		})
	})
	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf"})
	})
	mux.HandleFunc("/api/persons/7/groups", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "ChurchTools Admin"},
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := churchtools.NewClient(server.URL, "token", "", "")
	if err := client.Login(); err != nil {
		t.Fatal(err)
	}

	results, err := client.EnsurePreJoinGroups([]string{"ChurchTools Admin"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !results[0].Skipped {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestEnsurePreJoinGroupsRetriesHiddenGroupAfterEarlierJoin(t *testing.T) {
	adminVisible := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": 7, "firstName": "Admin", "lastName": "User", "email": "admin@example.org"},
		})
	})
	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf"})
	})
	mux.HandleFunc("/api/persons/7/groups", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})
	mux.HandleFunc("/api/groups", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		var data []map[string]any
		switch query {
		case "ChurchTools Admin":
			if adminVisible {
				data = []map[string]any{{"id": 1, "name": "ChurchTools Admin"}}
			}
		case "ChurchTools Verwaltung":
			data = []map[string]any{{"id": 2, "name": "ChurchTools Verwaltung"}}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
	})
	mux.HandleFunc("/api/groups/2/members/7", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		adminVisible = true
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/groups/1/members/7", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := churchtools.NewClient(server.URL, "token", "", "")
	if err := client.Login(); err != nil {
		t.Fatal(err)
	}

	results, err := client.EnsurePreJoinGroups([]string{"ChurchTools Admin", "ChurchTools Verwaltung"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %+v", results)
	}
	if results[0].GroupName != "ChurchTools Admin" || results[0].Status != churchtools.MembershipActive {
		t.Fatalf("admin result = %+v", results[0])
	}
	if results[1].GroupName != "ChurchTools Verwaltung" || results[1].Status != churchtools.MembershipActive {
		t.Fatalf("verwaltung result = %+v", results[1])
	}
}
