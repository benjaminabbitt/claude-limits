// Package auth handles loading Claude Code OAuth credentials.
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Credentials represents the OAuth credentials from Claude Code.
type Credentials struct {
	AccessToken      string
	RefreshToken     string
	ExpiresAt        time.Time
	SubscriptionType string
	RateLimitTier    string
}

// credentialsFile represents the JSON structure of ~/.claude/.credentials.json
type credentialsFile struct {
	ClaudeAiOauth struct {
		AccessToken      string   `json:"accessToken"`
		RefreshToken     string   `json:"refreshToken"`
		ExpiresAt        int64    `json:"expiresAt"`
		Scopes           []string `json:"scopes"`
		SubscriptionType string   `json:"subscriptionType"`
		RateLimitTier    string   `json:"rateLimitTier"`
	} `json:"claudeAiOauth"`
}

// DefaultCredentialsPath returns the default path to Claude Code credentials.
func DefaultCredentialsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", ".credentials.json")
}

// Load reads OAuth credentials from the specified path.
// If path is empty, uses the default Claude Code credentials path.
func Load(path string) (*Credentials, error) {
	if path == "" {
		path = DefaultCredentialsPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Claude Code credentials not found at %s - please authenticate with Claude Code first", path)
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var cf credentialsFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	if cf.ClaudeAiOauth.AccessToken == "" {
		return nil, fmt.Errorf("no OAuth access token found in credentials file")
	}

	return &Credentials{
		AccessToken:      cf.ClaudeAiOauth.AccessToken,
		RefreshToken:     cf.ClaudeAiOauth.RefreshToken,
		ExpiresAt:        time.UnixMilli(cf.ClaudeAiOauth.ExpiresAt),
		SubscriptionType: cf.ClaudeAiOauth.SubscriptionType,
		RateLimitTier:    cf.ClaudeAiOauth.RateLimitTier,
	}, nil
}

// IsExpired returns true if the access token has expired.
func (c *Credentials) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}
