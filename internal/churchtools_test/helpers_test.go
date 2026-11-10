package churchtools_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

// newLoggedInClient starts a mock ChurchTools server with whoami/csrftoken
// handlers added to mux and returns a logged-in token client.
func newLoggedInClient(t *testing.T, mux *http.ServeMux) *churchtools.Client {
	t.Helper()
	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": 1, "firstName": "A", "lastName": "B", "email": "a@example.org"},
		})
	})
	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf"})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := churchtools.NewClient(server.URL, "token", "", "")
	if err := client.Login(); err != nil {
		t.Fatal(err)
	}
	return client
}
