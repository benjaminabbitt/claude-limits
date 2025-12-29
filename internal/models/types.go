package models

import (
	"encoding/json"
)

// Usage represents the usage data from Claude.ai API.
// The API structure varies, so we preserve the raw JSON and parse flexibly.
type Usage struct {
	// Raw JSON response for output and inspection
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON captures the raw JSON for later use
func (u *Usage) UnmarshalJSON(data []byte) error {
	u.Raw = make(json.RawMessage, len(data))
	copy(u.Raw, data)
	return nil
}

// ToJSON returns the usage as a formatted JSON string
func (u *Usage) ToJSON() (string, error) {
	if u.Raw == nil {
		return "{}", nil
	}
	var formatted interface{}
	if err := json.Unmarshal(u.Raw, &formatted); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(formatted, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
