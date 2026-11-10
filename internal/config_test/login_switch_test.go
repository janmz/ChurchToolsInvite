package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	config "github.com/janmz/churchtools-invite/internal/config"
)

// CR-1: switching the login method must also drop the encrypted copy of the
// previous secret, otherwise sconfig decrypts it again on the next load.
func TestUsePasswordRemovesEncryptedToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"base_url":"https://x.church.tools","login_token":"old-token"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadFile(path) // encrypts old-token on disk
	if err != nil {
		t.Fatal(err)
	}

	cfg.UsePassword("user", "neues passwort")
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.LoginTokenPassword != "" {
		t.Fatalf("old token came back: %q", loaded.LoginTokenPassword)
	}
	if loaded.Username != "user" || loaded.Password != "neues passwort" {
		t.Fatalf("credentials = %q / %q", loaded.Username, loaded.Password)
	}
}

func TestUseLoginTokenRemovesEncryptedPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"base_url":"https://x.church.tools","username":"u","password":"old-pass"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	cfg.UseLoginToken(" new-token ")
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Password != "" || loaded.Username != "" || loaded.LoginTokenPassword != "new-token" {
		t.Fatalf("loaded = user %q, password %q, token %q", loaded.Username, loaded.Password, loaded.LoginTokenPassword)
	}
}

// CR-23: defaults are applied at runtime, not written to the file.
func TestSaveDoesNotWriteDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Default()
	cfg.BaseURL = "https://x.church.tools"
	cfg.UseLoginToken("token")
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"delay_ms", "pre_join_groups"} {
		if strings.Contains(string(data), key) {
			t.Fatalf("%s written to file:\n%s", key, data)
		}
	}
	if got := cfg.EffectiveDelayMS(); got != config.DefaultDelayMS {
		t.Fatalf("EffectiveDelayMS = %d, want %d", got, config.DefaultDelayMS)
	}
	if len(cfg.PreJoinGroupNames()) == 0 {
		t.Fatal("default pre-join groups missing at runtime")
	}
}

// CR-18: delay_ms must stay within 0 … MaxDelayMS.
func TestValidateRejectsDelayOutOfRange(t *testing.T) {
	for _, ms := range []int{-1, config.MaxDelayMS + 1} {
		cfg := config.Config{BaseURL: "https://x.church.tools", LoginTokenPassword: "t", DelayMS: ms}
		if err := cfg.Validate(); err == nil {
			t.Fatalf("delay_ms %d accepted", ms)
		}
	}
	cfg := config.Config{BaseURL: "https://x.church.tools", LoginTokenPassword: "t", DelayMS: 1000}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("delay_ms 1000 rejected: %v", err)
	}
}
