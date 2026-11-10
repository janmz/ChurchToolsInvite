package cmd

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
)

// S3: groups joined in a run are recorded; leftovers of an aborted run are
// offered for leaving on the next start; the file disappears when empty.
func TestTrackTemporaryGroups(t *testing.T) {
	dir := t.TempDir()
	origPath, origReader := configPath, stdinReader
	t.Cleanup(func() { configPath, stdinReader = origPath, origReader })
	configPath = filepath.Join(dir, "config.json")

	deleted := map[string]bool{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": 1}})
	})
	mux.HandleFunc("/api/csrftoken", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": "csrf"})
	})
	mux.HandleFunc("/api/groups/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			deleted[r.URL.Path] = true
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			http.NotFound(w, r) // not a member yet
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}})
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := churchtools.NewClient(server.URL, "token", "", "")
	if err := client.Login(); err != nil {
		t.Fatal(err)
	}

	// Leftover from an aborted earlier run.
	statePath := groupStatePath()
	if err := saveGroupState(statePath, groupState{Entries: []groupStateEntry{{BaseURL: client.BaseURL(), PersonID: 1, GroupIDs: []int{5}}}}); err != nil {
		t.Fatal(err)
	}

	stdinReader = bufio.NewReader(strings.NewReader("j\n"))
	trackTemporaryGroups(client)
	if !deleted["/api/groups/5/members/1"] {
		t.Fatal("leftover group 5 was not left")
	}

	if _, err := client.RequestGroupMembership(9, 1); err != nil {
		t.Fatal(err)
	}
	state, err := loadGroupState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := state.entry(client.BaseURL(), 1); len(got) != 1 || got[0] != 9 {
		t.Fatalf("recorded groups = %v, want [9]", got)
	}

	client.LeaveJoinedGroups()
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Fatalf("state file should be removed after leaving, stat err = %v", err)
	}
}
