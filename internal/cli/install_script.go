package cli

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sort"

	"github.com/benjaminabbitt/claude-limits/internal/claudecode"
	"github.com/benjaminabbitt/claude-limits/internal/scripts"

	"github.com/spf13/cobra"
)

var (
	forceOverwrite bool
	listScripts    bool
	projectSettings bool
)

var installScriptCmd = &cobra.Command{
	Use:   "install-script <name> <path>",
	Short: "Install an embedded script to a file path",
	Long: `Install one of the embedded status line scripts to a specified location.

This command:
1. Installs the script file to the specified path
2. Configures Claude Code's statusLine setting to use the script

Available scripts:
  bash        - Bash status line script for Claude Code
  powershell  - PowerShell status line script for Claude Code

The bash script will be installed with executable permissions (0755) on Unix systems.

By default, the statusLine is configured in user settings (~/.claude/settings.json).
Use --project to configure in project settings (.claude/settings.json) instead.

If statusLine is already configured, use --force to overwrite it.

Examples:
  claude-limits install-script bash ~/.local/bin/claude-limits-statusline.sh
  claude-limits install-script powershell ~/bin/claude-limits-statusline.ps1
  claude-limits install-script --project bash .local/bin/claude-limits-statusline.sh
  claude-limits install-script --list`,
	RunE: runInstallScript,
	Args: func(cmd *cobra.Command, args []string) error {
		if listScripts {
			return nil
		}
		if len(args) != 2 {
			return fmt.Errorf("requires exactly 2 arguments: <name> <path>")
		}
		return nil
	},
}

func init() {
	installScriptCmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing file and statusLine config")
	installScriptCmd.Flags().BoolVar(&listScripts, "list", false, "List available scripts")
	installScriptCmd.Flags().BoolVar(&projectSettings, "project", false, "Configure statusLine in project settings (.claude/settings.json)")
}

func runInstallScript(cmd *cobra.Command, args []string) error {
	if listScripts {
		return printAvailableScripts()
	}

	name := args[0]
	path := args[1]

	script := scripts.Get(name)
	if script == nil {
		return fmt.Errorf("unknown script: %s\nRun 'claude-limits install-script --list' to see available scripts", name)
	}

	// Check if file exists
	if _, err := os.Stat(path); err == nil {
		if !forceOverwrite {
			return fmt.Errorf("file already exists: %s\nUse --force to overwrite", path)
		}
	}

	// Check statusLine conflict before writing any files
	if err := checkStatusLineConflict(); err != nil {
		return err
	}

	// Determine permissions
	perm := os.FileMode(0644)
	if name == "bash" && runtime.GOOS != "windows" {
		perm = 0755
	}

	// Write the script file
	if err := os.WriteFile(path, script.Content, perm); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	fmt.Printf("Installed %s to %s\n", script.Filename, path)

	// Configure statusLine in Claude Code settings
	if err := configureStatusLine(path); err != nil {
		return err
	}

	return nil
}

func checkStatusLineConflict() error {
	if forceOverwrite {
		return nil
	}

	var settingsPath string
	var settingsType string

	if projectSettings {
		settingsPath = claudecode.DefaultProjectSettingsPath()
		settingsType = "project"
	} else {
		settingsPath = claudecode.DefaultUserSettingsPath()
		settingsType = "user"
	}

	settings, err := claudecode.LoadSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load Claude Code settings: %w", err)
	}

	if settings.HasStatusLine() {
		return fmt.Errorf("statusLine already configured in %s settings (%s)\nUse --force to overwrite", settingsType, settingsPath)
	}

	return nil
}

func configureStatusLine(scriptPath string) error {
	var settingsPath string
	var settingsType string

	if projectSettings {
		settingsPath = claudecode.DefaultProjectSettingsPath()
		settingsType = "project"
	} else {
		settingsPath = claudecode.DefaultUserSettingsPath()
		settingsType = "user"
	}

	settings, err := claudecode.LoadSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load Claude Code settings: %w", err)
	}

	if err := settings.SetStatusLine(scriptPath, forceOverwrite); err != nil {
		if errors.Is(err, claudecode.ErrStatusLineExists) {
			return fmt.Errorf("statusLine already configured in %s settings (%s)\nUse --force to overwrite", settingsType, settingsPath)
		}
		return err
	}

	if err := claudecode.SaveSettings(settingsPath, settings); err != nil {
		return fmt.Errorf("failed to save Claude Code settings: %w", err)
	}

	fmt.Printf("Configured statusLine in %s settings (%s)\n", settingsType, settingsPath)
	return nil
}

func printAvailableScripts() error {
	fmt.Println("Available scripts:")
	fmt.Println()

	// Sort script names for consistent output
	names := scripts.List()
	sort.Strings(names)

	for _, name := range names {
		script := scripts.Get(name)
		fmt.Printf("  %-12s %s\n", name, script.Description)
	}

	fmt.Println()
	fmt.Println("Usage: claude-limits install-script <name> <path>")
	return nil
}
