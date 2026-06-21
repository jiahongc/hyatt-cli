// Copyright 2026 Jiahong Chen and contributors. Licensed under Apache-2.0. See LICENSE.
// Maintained in the World of Hyatt CLI repository.

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	BaseURL       string            `json:"base_url"`
	AuthHeaderVal string            `json:"auth_header"`
	Headers       map[string]string `json:"headers,omitempty"`
	AuthSource    string            `json:"-"`
	AccessToken   string            `json:"access_token"`
	RefreshToken  string            `json:"refresh_token"`
	TokenExpiry   time.Time         `json:"token_expiry"`
	ClientID      string            `json:"client_id"`
	ClientSecret  string            `json:"client_secret"`
	Path          string            `json:"-"`
	envOverrides  map[string]bool   `json:"-"`
	fileConfig    *Config           `json:"-"`
	HyattCookies  string            `json:"cookies"`
}

func Load(configPath string) (*Config, error) {
	cfg := &Config{
		BaseURL: "https://www.hyatt.com",
	}

	// Resolve config path
	path := configPath
	if path == "" {
		path = os.Getenv("HYATT_CONFIG")
	}
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".config", "hyatt-cli", "config.json")
	}
	cfg.Path = path

	// Try to load config file
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config %s: %w", path, err)
		}
		cfg.Path = path
	}

	cfg.snapshotFileConfig()

	// Env var overrides
	if v := os.Getenv("HYATT_COOKIES"); v != "" {
		cfg.HyattCookies = v
		cfg.markEnvOverride("HyattCookies")
		cfg.AuthSource = "env:HYATT_COOKIES"
	}

	// Label config-file-derived credentials so doctor can distinguish
	// "credentials persisted on disk" from "no credentials at all" — without
	// this, users who saved via set-token without an env var see a blank
	// auth_source and can't tell whether their config is being picked up.
	// The label is the literal "config" rather than "config:<path>"; the
	// config file path is exposed separately as report["config_path"], and
	// embedding it in auth_source leaks the user's home directory through
	// doctor's JSON envelope.
	if cfg.AuthSource == "" && (cfg.AuthHeaderVal != "" || cfg.AccessToken != "") {
		cfg.AuthSource = "config"
	}
	if cfg.AuthSource == "" && cfg.HyattCookies != "" {
		cfg.AuthSource = "config"
	}

	// Soft agentcookie integration: if the agentcookie daemon manages this
	// CLI's secrets, it writes a marker file alongside the config file. When
	// the marker is present AND credentials came from the config (not from a
	// direct env var override that wins above), upgrade AuthSource to
	// "agentcookie" so doctor / auth-status can surface the bus state. When
	// the marker is absent, behavior is identical to pre-agentcookie: no
	// import, no network, no error. agentcookie itself is never imported
	// here — the contract is purely on-disk.
	if cfg.AuthSource == "config" {
		marker := filepath.Join(filepath.Dir(cfg.Path), ".agentcookie-managed")
		if _, err := os.Stat(marker); err == nil {
			cfg.AuthSource = "agentcookie"
		}
	}

	// Base URL override used by verification to point at mock/test servers.
	if v := os.Getenv("HYATT_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	return cfg, nil
}

func (c *Config) AuthHeader() string {
	if c.AuthHeaderVal != "" {
		return c.AuthHeaderVal
	}
	// Browser cookies are already a complete Cookie header value. Do not wrap
	// them in an auth scheme; callers must send them on the Cookie header.
	if c.HyattCookies != "" {
		if c.AuthSource == "" {
			c.AuthSource = "env:HYATT_COOKIES"
		}
		return c.HyattCookies
	}
	if c.AccessToken != "" {
		if c.AuthSource == "" {
			c.AuthSource = "browser"
		}
		return ensureAuthScheme("Bearer", c.AccessToken)
	}
	return ""
}

func (c *Config) UsesCookieAuth() bool {
	return c != nil && c.AuthHeaderVal == "" && c.HyattCookies != ""
}

func applyAuthFormat(format string, replacements map[string]string) string {
	if format == "" {
		return ""
	}
	for key, value := range replacements {
		format = strings.ReplaceAll(format, "{"+key+"}", value)
	}
	if strings.Contains(format, "{") {
		return ""
	}
	return format
}

// ensureAuthScheme returns "<scheme> <token>" but skips the prefix when the
// token already carries it case-insensitively, so a user who exports the
// env var with the scheme already attached doesn't end up double-prefixed.
// Empty scheme returns the token as-is.
func ensureAuthScheme(scheme, token string) string {
	if token == "" {
		return ""
	}
	if scheme == "" {
		return token
	}
	schemeWithSpace := scheme + " "
	if len(token) >= len(schemeWithSpace) && strings.EqualFold(token[:len(schemeWithSpace)], schemeWithSpace) {
		return token
	}
	return schemeWithSpace + token
}

func (c *Config) SaveTokens(clientID, clientSecret, accessToken, refreshToken string, expiry time.Time) error {
	c.ClientID = clientID
	c.ClientSecret = clientSecret
	c.AccessToken = accessToken
	c.RefreshToken = refreshToken
	c.TokenExpiry = expiry
	delete(c.envOverrides, "ClientID")
	delete(c.envOverrides, "ClientSecret")
	delete(c.envOverrides, "AccessToken")
	delete(c.envOverrides, "RefreshToken")
	delete(c.envOverrides, "TokenExpiry")
	c.updateFileConfigField("ClientID")
	c.updateFileConfigField("ClientSecret")
	c.updateFileConfigField("AccessToken")
	c.updateFileConfigField("RefreshToken")
	c.updateFileConfigField("TokenExpiry")
	return c.save()
}

func (c *Config) ClearTokens() error {
	// AuthHeader() falls back to the env-var-derived fields when AuthHeaderVal
	// and AccessToken are empty, so dropping the working credential requires
	// zeroing every emitted credential field, not just the OAuth trio.
	// ClientID/ClientSecret persist to disk via SaveTokens for the oauth2
	// and oauth2-cc flows, so logout must wipe them too; otherwise
	// `auth login` can re-mint a new access token unattended.
	c.AuthHeaderVal = ""
	c.AccessToken = ""
	c.RefreshToken = ""
	c.TokenExpiry = time.Time{}
	c.ClientID = ""
	c.ClientSecret = ""
	delete(c.envOverrides, "AuthHeaderVal")
	delete(c.envOverrides, "AccessToken")
	delete(c.envOverrides, "RefreshToken")
	delete(c.envOverrides, "TokenExpiry")
	delete(c.envOverrides, "ClientID")
	delete(c.envOverrides, "ClientSecret")
	c.updateFileConfigField("AuthHeaderVal")
	c.updateFileConfigField("AccessToken")
	c.updateFileConfigField("RefreshToken")
	c.updateFileConfigField("TokenExpiry")
	c.updateFileConfigField("ClientID")
	c.updateFileConfigField("ClientSecret")
	c.HyattCookies = ""
	delete(c.envOverrides, "HyattCookies")
	return c.save()
}

func (c *Config) markEnvOverride(field string) {
	if c.envOverrides == nil {
		c.envOverrides = map[string]bool{}
	}
	c.envOverrides[field] = true
}

// cloneStringMap returns an independent copy of m (nil stays nil). The fileConfig
// snapshot must not share reference-type map fields (such as Headers) with the
// live config, or a later mutation to one would silently track in the other.
func cloneStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (c *Config) snapshotFileConfig() {
	snapshot := *c
	snapshot.envOverrides = nil
	snapshot.fileConfig = nil
	// *c is a shallow copy: map fields are reference types, so the snapshot would
	// share them with c and silently track later mutations, defeating the
	// isolation this snapshot exists to provide. Clone them.
	snapshot.Headers = cloneStringMap(c.Headers)
	c.fileConfig = &snapshot
}

func (c *Config) configForSave() Config {
	out := *c
	if c.fileConfig != nil {
		if c.envOverrides["HyattCookies"] {
			out.HyattCookies = c.fileConfig.HyattCookies
		}
	}
	out.envOverrides = nil
	out.fileConfig = nil
	return out
}

func (c *Config) updateFileConfigField(field string) {
	if c.fileConfig == nil || c.envOverrides[field] {
		return
	}
	switch field {
	case "AuthHeaderVal":
		c.fileConfig.AuthHeaderVal = c.AuthHeaderVal
	case "AccessToken":
		c.fileConfig.AccessToken = c.AccessToken
	case "RefreshToken":
		c.fileConfig.RefreshToken = c.RefreshToken
	case "TokenExpiry":
		c.fileConfig.TokenExpiry = c.TokenExpiry
	case "ClientID":
		c.fileConfig.ClientID = c.ClientID
	case "ClientSecret":
		c.fileConfig.ClientSecret = c.ClientSecret
	case "HyattCookies":
		c.fileConfig.HyattCookies = c.HyattCookies
	}
}

func (c *Config) save() error {
	dir := filepath.Dir(c.Path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	persisted := c.configForSave()
	data, err := json.MarshalIndent(persisted, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(c.Path, data, 0o600); err != nil {
		return err
	}
	c.fileConfig = &persisted
	c.fileConfig.envOverrides = nil
	c.fileConfig.fileConfig = nil
	// persisted shares its map fields with c (configForSave shallow-copies *c),
	// so isolate the stored fileConfig the same way snapshotFileConfig does;
	// otherwise later mutations to c's maps leak into the on-disk snapshot.
	c.fileConfig.Headers = cloneStringMap(c.fileConfig.Headers)
	return nil
}

// Ensure strings import is used
var _ = strings.ReplaceAll
