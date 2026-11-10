package cmd

import (
	"bufio"
	"errors"
	"strings"
	"testing"
)

// CR-2: "n" or "nein" must never become https://n.church.tools.
func TestAskMainInstanceURL(t *testing.T) {
	orig, origDiscover := stdinReader, discoverMainInstanceURL
	t.Cleanup(func() { stdinReader, discoverMainInstanceURL = orig, origDiscover })
	// The sub-instance reports its main instance via the OAuth redirect.
	discoverMainInstanceURL = func(string) (string, error) { return "https://gemeinde.church.tools", nil }

	const base = "https://gemeinde-unterstadt.church.tools"
	cases := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"\n", "", false},
		{"n\n", "", false},
		{"Nein\n", "", false},
		{"-\n", "", false},
		{"j\n", "https://gemeinde.church.tools", false},
		{"andere-gemeinde\n", "https://andere-gemeinde.church.tools", false},
		{"x\n", "", true},
		{"http://evil.example.org\n", "", true},
		{base + "\n", "", true},
	}
	for _, tc := range cases {
		stdinReader = bufio.NewReader(strings.NewReader(tc.input))
		got, err := askMainInstanceURL(base, "")
		if (err != nil) != tc.wantErr || got != tc.want {
			t.Errorf("input %q: got %q, err %v; want %q, err %v", tc.input, got, err, tc.want, tc.wantErr)
		}
	}

	// S8: without a reported main instance the guess from the name must be
	// typed in; "j" is not accepted.
	discoverMainInstanceURL = func(string) (string, error) { return "", errors.New("kein oauth") }
	stdinReader = bufio.NewReader(strings.NewReader("j\n"))
	if got, err := askMainInstanceURL(base, ""); err == nil || got != "" {
		t.Fatalf("guess accepted with j: got %q, err %v", got, err)
	}
}
