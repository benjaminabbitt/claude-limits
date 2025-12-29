package cli

import (
	"fmt"
	"os"
	"runtime"
	"sort"

	"github.com/benjaminabbitt/claude-limits/internal/scripts"

	"github.com/spf13/cobra"
)

var (
	forceOverwrite bool
	listScripts    bool
)

var installScriptCmd = &cobra.Command{
	Use:   "install-script <name> <path>",
	Short: "Install an embedded script to a file path",
	Long: `Install one of the embedded status line scripts to a specified location.

Available scripts:
  bash        - Bash status line script for Claude Code
  powershell  - PowerShell status line script for Claude Code

The bash script will be installed with executable permissions (0755) on Unix systems.

Examples:
  claude-limits install-script bash ~/.local/bin/claude-limits-statusline.sh
  claude-limits install-script powershell ~/bin/claude-limits-statusline.ps1
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
	installScriptCmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing file")
	installScriptCmd.Flags().BoolVar(&listScripts, "list", false, "List available scripts")
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

	// Determine permissions
	perm := os.FileMode(0644)
	if name == "bash" && runtime.GOOS != "windows" {
		perm = 0755
	}

	// Write the file
	if err := os.WriteFile(path, script.Content, perm); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	fmt.Printf("Installed %s to %s\n", script.Filename, path)
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
