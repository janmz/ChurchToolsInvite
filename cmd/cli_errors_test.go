package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestTranslateCLIError(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantSame bool
	}{
		{
			name: "unknown flag",
			in:   "unknown flag: --foo",
			want: "unbekannte Option --foo",
		},
		{
			name: "unknown command",
			in:   `unknown command "foo" for "Churchtools-Invite"`,
			want: `unbekannter Befehl "foo" (Programm "Churchtools-Invite")`,
		},
		{
			name: "required csv",
			in:   `required flag(s) "csv" not set`,
			want: "Pflichtoption --csv (-f) fehlt",
		},
		{
			name: "required generic",
			in:   `required flag(s) "output" not set`,
			want: "Pflichtoption --output fehlt",
		},
		{
			name: "invalid int",
			in:   `invalid argument "abc" for "--status-id" flag: strconv.ParseInt: parsing "abc": invalid syntax`,
			want: `ungültiger Wert "abc" für Option --status-id: keine Zahl`,
		},
		{
			name:     "passthrough",
			in:       "csv öffnen: datei fehlt",
			wantSame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateCLIError(errors.New(tt.in)).Error()
			if tt.wantSame {
				if got != tt.in {
					t.Fatalf("got %q want same %q", got, tt.in)
				}
				return
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestFinalizeCLIErrorUsageError(t *testing.T) {
	inner := errors.New("Pflichtoption --csv (-f) fehlt")
	err := finalizeCLIError(&usageError{cmd: inviteCmd, err: inner})
	if err == nil || err.Error() != inner.Error() {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetupSubcommandArgs(t *testing.T) {
	err := setupSubcommandArgs(setupCmd, []string{"foo"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unbekannter Unterbefehl") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestSetupSubcommandRequired(t *testing.T) {
	err := setupSubcommandRequired(setupCmd, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "bitte einen Unterbefehl") {
		t.Fatalf("unexpected message: %v", err)
	}
}
