package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/janmz/churchtools-invite/internal/fileacl"
	"github.com/janmz/churchtools-invite/internal/iolimit"
)

const DefaultConfigName = "config.json"

// Config holds ChurchTools connection settings.
//
// Secrets are stored with github.com/janmz/sconfig: a plaintext value in
// "login_token" or "password" is encrypted (machine-bound) into
// "login_token_secure" / "secure_password" on the first load and replaced by
// a marker. In memory the fields hold the decrypted values.
type Config struct {
	BaseURL string `json:"base_url"`
	// MainURL is the main (central) instance for a sub-instance login. It is
	// only used when set explicitly; no host is derived automatically.
	MainURL                  string `json:"main_url,omitempty"`
	LoginTokenPassword       string `json:"login_token,omitempty"`
	LoginTokenSecurePassword string `json:"login_token_secure,omitempty"`
	Username                 string `json:"username,omitempty"`
	Password                 string `json:"password,omitempty"`
	SecurePassword           string `json:"secure_password,omitempty"`
	DelayMS                  int    `json:"delay_ms,omitempty"`
	CampusID                 int    `json:"campus_id,omitempty"`
	PreJoinGroups            string `json:"pre_join_groups,omitempty"`
}

// Central defaults. Change them only here.
const (
	// DefaultDelayMS is the pause after each write request of invite
	// (invitation, e-mail update) in milliseconds, used when delay_ms is 0.
	DefaultDelayMS = 500
	// MaxDelayMS is the largest accepted delay_ms / --delay-ms value.
	MaxDelayMS = 60_000
)

// Default returns an empty configuration. Defaults such as DefaultDelayMS
// and DefaultPreJoinGroups are applied at runtime (EffectiveDelayMS,
// PreJoinGroupNames) and never written to the file, so a changed default in
// a later version reaches existing installations.
func Default() Config {
	return Config{}
}

// EffectiveDelayMS returns delay_ms, or DefaultDelayMS when it is not set.
func (c Config) EffectiveDelayMS() int {
	if c.DelayMS <= 0 {
		return DefaultDelayMS
	}
	return c.DelayMS
}

// ValidateDelayMS checks a delay value from the config or the command line.
func ValidateDelayMS(field string, ms int) error {
	if ms < 0 || ms > MaxDelayMS {
		return fmt.Errorf("%s muss zwischen 0 und %d ms liegen (ist %d)", field, MaxDelayMS, ms)
	}
	return nil
}

// UseLoginToken switches the login to token and removes stored passwords,
// including the encrypted copies (otherwise sconfig would decrypt them again
// on the next load and the old credentials would stay active).
func (c *Config) UseLoginToken(token string) {
	c.LoginTokenPassword = strings.TrimSpace(token)
	c.LoginTokenSecurePassword = ""
	c.Username = ""
	c.Password = ""
	c.SecurePassword = ""
}

// UsePassword switches the login to username/password and removes a stored
// login token, including its encrypted copy. The password is kept verbatim.
func (c *Config) UsePassword(username, password string) {
	c.Username = strings.TrimSpace(username)
	c.Password = password
	c.SecurePassword = ""
	c.LoginTokenPassword = ""
	c.LoginTokenSecurePassword = ""
}

// DefaultEditPersonsGroups are candidate groups for write/admin permissions.
var DefaultEditPersonsGroups = []string{
	"Personen Administration",
	"Personen bearbeiten",
}

// DefaultExportPersonsGroups are candidate groups for export permission.
var DefaultExportPersonsGroups = []string{
	"Personen exportieren",
}

// DefaultPreJoinGroups is the comma-separated list of groups joined before export/invite.
const DefaultPreJoinGroups = "ChurchTools Verwaltung,Gruppen Administration,ChurchTools Admin,Personen Administration,Personen verwalten"

// ParseCommaSeparatedNames splits a comma-separated list and trims empty entries.
func ParseCommaSeparatedNames(raw string) []string {
	parts := strings.Split(raw, ",")
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		names = append(names, part)
	}
	return names
}

// PreJoinGroupNames returns groups to join in order before export/invite.
// Set pre_join_groups to "-" or "none" to disable.
func (c Config) PreJoinGroupNames() []string {
	raw := strings.TrimSpace(c.PreJoinGroups)
	if raw == "-" || strings.EqualFold(raw, "none") {
		return nil
	}
	if raw == "" {
		return ParseCommaSeparatedNames(DefaultPreJoinGroups)
	}
	return ParseCommaSeparatedNames(raw)
}

// Load reads configuration from a JSON file and applies environment overrides.
func Load(path string) (Config, error) {
	cfg, err := LoadFile(path)
	if err != nil {
		return cfg, err
	}

	cfg.applyEnv()
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// LoadFile reads only the JSON file (no environment overrides, no
// validation, no defaults filled in). Plaintext secrets in the file are encrypted on
// disk by sconfig; the returned struct holds the decrypted values. Use it
// when the file is to be modified and saved again, so values from the
// environment do not end up on disk.
func LoadFile(path string) (Config, error) {
	cfg := Default()
	// Same limit and error text as for CSV files, before sconfig parses it.
	if _, err := iolimit.ReadFile(path, iolimit.MaxConfigBytes); err != nil {
		return cfg, fmt.Errorf("config lesen (%s): %w", path, err)
	}
	if err := loadSecure(&cfg, path); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// LoadOrEmpty returns config from file when present, otherwise environment only.
func LoadOrEmpty(path string) (Config, error) {
	if path == "" {
		path = DefaultConfigName
	}

	if !Exists(path) {
		cfg := Default()
		cfg.applyEnv()
		return cfg, cfg.Validate()
	}

	return Load(path)
}

// Exists reports whether a config file exists at path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// SaveCampusID stores campus_id in the config file without touching any other
// key. Unknown keys are preserved and environment values are never written.
// A missing file is created.
func SaveCampusID(path string, campusID int) error {
	if path == "" {
		path = DefaultConfigName
	}

	raw := map[string]any{}
	if Exists(path) {
		data, err := iolimit.ReadFile(path, iolimit.MaxConfigBytes)
		if err != nil {
			return fmt.Errorf("config lesen (%s): %w", path, err)
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("config parsen: %w", err)
		}
	}
	raw["campus_id"] = campusID
	return writeJSON(path, raw)
}

// Save writes the configuration with encrypted secrets and restricts the
// file to the current user.
func Save(path string, cfg Config) error {
	if path == "" {
		path = DefaultConfigName
	}
	return saveSecure(&cfg, path)
}

// writeJSON writes value as indented JSON, atomically and restricted to the
// current user.
func writeJSON(path string, value any) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("config-verzeichnis anlegen: %w", err)
		}
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("config serialisieren: %w", err)
	}

	data = append(data, '\n')

	// Atomic: a failed write must not leave an empty file (the encrypted
	// credentials would be lost).
	if err := fileacl.WriteAtomic(path, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	}); err != nil {
		return fmt.Errorf("config speichern: %w", err)
	}
	return nil
}

func (c *Config) applyEnv() {
	if v := strings.TrimSpace(os.Getenv("CT_BASE_URL")); v != "" {
		c.BaseURL = v
	}
	if v := strings.TrimSpace(os.Getenv("CT_MAIN_URL")); v != "" {
		c.MainURL = v
	}
	if v := strings.TrimSpace(os.Getenv("CT_LOGIN_TOKEN")); v != "" {
		c.LoginTokenPassword = v
	}
	if v := strings.TrimSpace(os.Getenv("CT_USERNAME")); v != "" {
		c.Username = v
	}
	// The password is taken verbatim: spaces may be part of it.
	if v := os.Getenv("CT_PASSWORD"); strings.TrimSpace(v) != "" {
		c.Password = v
	}
	if v := strings.TrimSpace(os.Getenv("CT_PRE_JOIN_GROUPS")); v != "" {
		c.PreJoinGroups = v
	}
}

// Validate checks required fields.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.BaseURL) == "" {
		return errors.New("base_url fehlt (config oder CT_BASE_URL)")
	}

	c.BaseURL = NormalizeBaseURL(c.BaseURL)
	if err := validateInstanceURL("base_url", c.BaseURL); err != nil {
		return err
	}
	if strings.TrimSpace(c.MainURL) != "" {
		c.MainURL = NormalizeBaseURL(c.MainURL)
		if err := validateInstanceURL("main_url", c.MainURL); err != nil {
			return err
		}
	}

	if err := ValidateDelayMS("delay_ms", c.DelayMS); err != nil {
		return err
	}

	hasToken := strings.TrimSpace(c.LoginTokenPassword) != ""
	hasPassword := strings.TrimSpace(c.Username) != "" && strings.TrimSpace(c.Password) != ""
	if !hasToken && !hasPassword {
		return errors.New("login_token oder username/password erforderlich")
	}

	return nil
}

// NormalizeBaseURL removes trailing slashes and /api suffixes. When raw is only
// an instance name (no scheme), https://{name}.church.tools is assumed.
func NormalizeBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		if url, err := BaseURLFromInstanceName(raw); err == nil {
			return url
		}
	}
	url := raw
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, "/api")
	return url
}
