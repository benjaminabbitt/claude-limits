package scripts

import (
	_ "embed"
)

//go:embed claude-limits-statusline.sh
var bashScript []byte

//go:embed claude-limits-statusline.ps1
var powershellScript []byte

// Script represents an embedded script
type Script struct {
	Name        string
	Filename    string
	Description string
	Content     []byte
}

// Available scripts
var Available = map[string]Script{
	"bash": {
		Name:        "bash",
		Filename:    "claude-limits-statusline.sh",
		Description: "Bash status line script for Claude Code",
		Content:     bashScript,
	},
	"powershell": {
		Name:        "powershell",
		Filename:    "claude-limits-statusline.ps1",
		Description: "PowerShell status line script for Claude Code",
		Content:     powershellScript,
	},
}

// Get returns a script by name, or nil if not found
func Get(name string) *Script {
	if s, ok := Available[name]; ok {
		return &s
	}
	return nil
}

// List returns all available script names
func List() []string {
	names := make([]string, 0, len(Available))
	for name := range Available {
		names = append(names, name)
	}
	return names
}
