package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	churchtools "github.com/janmz/churchtools-invite/internal/churchtools"
	config "github.com/janmz/churchtools-invite/internal/config"
	"github.com/janmz/churchtools-invite/internal/fileacl"
	"github.com/janmz/churchtools-invite/internal/iolimit"
)

// groupStateFile records the groups the tool joined temporarily, next to the
// config file. If a run ends without leaving them (crash, closed window,
// second Ctrl+C, network error), the next start offers to leave them.
const groupStateFile = ".churchtools-invite-groups.json"

type groupStateEntry struct {
	BaseURL  string `json:"base_url"`
	PersonID int    `json:"person_id"`
	GroupIDs []int  `json:"group_ids"`
}

type groupState struct {
	Entries []groupStateEntry `json:"entries"`
}

func groupStatePath() string {
	path := configPath
	if path == "" {
		path = config.DefaultConfigName
	}
	return filepath.Join(filepath.Dir(path), groupStateFile)
}

func loadGroupState(path string) (groupState, error) {
	var state groupState
	data, err := iolimit.ReadFile(path, iolimit.MaxConfigBytes)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("%s parsen: %w", path, err)
	}
	return state, nil
}

func saveGroupState(path string, state groupState) error {
	entries := make([]groupStateEntry, 0, len(state.Entries))
	for _, entry := range state.Entries {
		if len(entry.GroupIDs) > 0 {
			entries = append(entries, entry)
		}
	}
	if len(entries) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	data, err := json.MarshalIndent(groupState{Entries: entries}, "", "  ")
	if err != nil {
		return err
	}
	return fileacl.WriteAtomic(path, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	})
}

// setEntry replaces the group list for baseURL/personID.
func (s *groupState) setEntry(baseURL string, personID int, groupIDs []int) {
	for i := range s.Entries {
		if strings.EqualFold(s.Entries[i].BaseURL, baseURL) && s.Entries[i].PersonID == personID {
			s.Entries[i].GroupIDs = groupIDs
			return
		}
	}
	s.Entries = append(s.Entries, groupStateEntry{BaseURL: baseURL, PersonID: personID, GroupIDs: groupIDs})
}

func (s groupState) entry(baseURL string, personID int) []int {
	for _, entry := range s.Entries {
		if strings.EqualFold(entry.BaseURL, baseURL) && entry.PersonID == personID {
			return entry.GroupIDs
		}
	}
	return nil
}

// trackTemporaryGroups offers to leave groups left over from an earlier run
// and then records every group joined in this run in the state file.
func trackTemporaryGroups(client *churchtools.Client) {
	path := groupStatePath()
	state, err := loadGroupState(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warnung: %s nicht lesbar (%v) – beigetretene Gruppen werden nicht protokolliert.\n", path, err)
		return
	}
	baseURL, personID := client.BaseURL(), client.PersonID()

	leftover := leaveLeftoverGroups(client, state.entry(baseURL, personID))
	persist := func(current []int) {
		state.setEntry(baseURL, personID, mergeGroupIDs(leftover, current))
		if err := saveGroupState(path, state); err != nil {
			fmt.Fprintf(os.Stderr, "Warnung: beigetretene Gruppen nicht in %s gespeichert: %v\n", path, err)
		}
	}
	persist(nil)
	client.SetJoinedGroupsObserver(persist)
}

// leaveLeftoverGroups asks whether groups from an aborted run should be left
// now and returns the groups that are still joined afterwards.
func leaveLeftoverGroups(client *churchtools.Client, groupIDs []int) []int {
	if len(groupIDs) == 0 {
		return nil
	}
	fmt.Fprintf(os.Stderr, "Hinweis: Bei einem früheren, abgebrochenen Lauf ist das Konto temporär Gruppen beigetreten, die nicht wieder verlassen wurden (Gruppen-IDs %s).\n", joinIDs(groupIDs))
	answer, err := readLine("Diese Gruppen jetzt verlassen? [J/n]: ")
	if err != nil || strings.EqualFold(strings.TrimSpace(answer), "n") {
		fmt.Fprintln(os.Stderr, "Gruppen bleiben bestehen; die Frage kommt beim nächsten Start erneut.")
		return groupIDs
	}

	var remaining []int
	for _, result := range client.LeaveGroups(groupIDs) {
		if result.Err != nil {
			fmt.Fprintf(os.Stderr, "Warnung: Gruppe ID %d konnte nicht verlassen werden: %v\n", result.GroupID, result.Err)
			remaining = append(remaining, result.GroupID)
			continue
		}
		fmt.Fprintf(os.Stderr, "Gruppe ID %d verlassen (Rest eines früheren Laufs).\n", result.GroupID)
	}
	return remaining
}

func mergeGroupIDs(a, b []int) []int {
	seen := make(map[int]bool, len(a)+len(b))
	var merged []int
	for _, id := range append(append([]int(nil), a...), b...) {
		if !seen[id] {
			seen[id] = true
			merged = append(merged, id)
		}
	}
	sort.Ints(merged)
	return merged
}

func joinIDs(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprint(id)
	}
	return strings.Join(parts, ", ")
}
