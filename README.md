# claude-limits

CLI tool to check your Claude.ai usage limits for Pro/Max subscriptions.

## Features

- Check 5-hour and weekly usage limits from the terminal
- Automatic authentication via browser cookies (Chrome, Firefox)
- MCP server mode for integration with Claude Code and other MCP clients
- Status line scripts for real-time usage in Claude Code
- Configurable date/time formats (12-hour, 24-hour, ISO 8601, etc.)
- Colorized output with utilization thresholds (green/yellow/red)
- Fuzzy field matching for quick queries
- Response caching to reduce API calls
- Cross-platform: Linux, macOS, Windows

## Installation

### Homebrew (macOS/Linux)

```bash
brew install benjaminabbitt/tap/claude-limits
```

### Go Install

```bash
go install github.com/benjaminabbitt/claude-limits/cmd/claude-limits@latest
```

### From Source

```bash
git clone https://github.com/benjaminabbitt/claude-limits.git
cd claude-limits
just build
```

## Usage

### Basic Usage

```bash
# Show all usage data (auto-detects browser cookies)
claude-limits

# Query specific field using fuzzy matching
claude-limits five          # Returns 5-hour utilization
claude-limits weekly        # Returns weekly utilization

# Output as JSON
claude-limits --format json
```

### Authentication

Authentication is resolved in this order:

1. CLI flags: `--cookie` and `--org-id`
2. Environment variables: `CLAUDE_SESSION_COOKIE` and `CLAUDE_ORG_ID`
3. Configuration file
4. Automatic extraction from browser cookies (Chrome, Firefox)

For headless environments, set the environment variables or use a config file:

```bash
export CLAUDE_SESSION_COOKIE="your-session-cookie"
export CLAUDE_ORG_ID="your-org-id"
```

## Configuration

Create a config file at `~/.config/claude-limits/config.yaml` (Linux/macOS) or `%APPDATA%\claude-limits\config.yaml` (Windows):

```yaml
# Authentication credentials
auth:
  session_cookie: "your-session-cookie"
  org_id: "your-org-id"

# Display formats using Go time layout syntax
# See: https://pkg.go.dev/time#pkg-constants
formats:
  # Use a preset: "12hour", "24hour", "iso8601", "us", or "eu"
  preset: "12hour"

  # Or specify custom formats (overrides preset):
  # datetime: "Mon, Jan 2 2006 at 3:04 PM MST"
  # date: "Mon, Jan 2 2006"
  # time: "3:04 PM"
```

### Format Presets

| Preset | Datetime | Date | Time |
|--------|----------|------|------|
| `12hour` (default) | `Mon, Jan 2 2006 at 3:04 PM MST` | `Mon, Jan 2 2006` | `3:04 PM` |
| `24hour` | `Mon, Jan 2 2006 at 15:04 MST` | `Mon, Jan 2 2006` | `15:04` |
| `iso8601` | `2006-01-02T15:04:05Z07:00` | `2006-01-02` | `15:04:05` |
| `us` | `Jan 2, 2006 3:04 PM MST` | `Jan 2, 2006` | `3:04 PM` |
| `eu` | `2 Jan 2006 15:04 MST` | `2 Jan 2006` | `15:04` |

Override the config file location with `--config` flag or `CLAUDE_LIMITS_CONFIG` env var.

### MCP Server

Run as an MCP server for integration with Claude Code or other MCP clients:

```bash
claude-limits serve
```

The server exposes a `get_usage` tool that returns current usage data as JSON.

#### Claude Code Configuration

Add to your Claude Code MCP settings:

```json
{
  "mcpServers": {
    "claude-limits": {
      "command": "claude-limits",
      "args": ["serve"]
    }
  }
}
```

### Status Line Integration

Install status line scripts for Claude Code:

```bash
# List available scripts
claude-limits install-script --list

# Install bash script
claude-limits install-script bash ~/.local/bin/claude-limits-statusline.sh

# Install PowerShell script
claude-limits install-script powershell ~/bin/claude-limits-statusline.ps1
```

Configure in Claude Code settings:

```json
{
  "statusLineCommand": "~/.local/bin/claude-limits-statusline.sh"
}
```

The status line shows: `5h: 45% @ 2:30 PM | wk: 23% @ Tue 8:00 AM | ctx: 67%`

#### Status Line Time Formats

Customize time formats via environment variables:

**Bash (strftime syntax):**
```bash
# 24-hour format
export CLAUDE_LIMITS_TIME_FORMAT="+%H:%M"
export CLAUDE_LIMITS_DATETIME_FORMAT="+%a %H:%M"

# 12-hour format (default)
export CLAUDE_LIMITS_TIME_FORMAT="+%I:%M %p"
export CLAUDE_LIMITS_DATETIME_FORMAT="+%a %I:%M %p"
```

**PowerShell (.NET format syntax):**
```powershell
# 24-hour format
$env:CLAUDE_LIMITS_TIME_FORMAT = "HH:mm"
$env:CLAUDE_LIMITS_DATETIME_FORMAT = "ddd HH:mm"

# 12-hour format (default)
$env:CLAUDE_LIMITS_TIME_FORMAT = "h:mm tt"
$env:CLAUDE_LIMITS_DATETIME_FORMAT = "ddd h:mm tt"
```

## Options

| Flag | Environment Variable | Description |
|------|---------------------|-------------|
| `--config` | `CLAUDE_LIMITS_CONFIG` | Config file path |
| `--cookie` | `CLAUDE_SESSION_COOKIE` | Claude.ai session cookie |
| `--org-id` | `CLAUDE_ORG_ID` | Claude.ai organization ID |
| `--format` | - | Output format: `table` (default) or `json` |
| `--cache` | - | Cache TTL in seconds (default: 30, 0 to disable) |
| `--no-color` | - | Disable colored output |
| `-v, --verbose` | - | Verbose output |

## Commands

| Command | Description |
|---------|-------------|
| `limits [query]` | Display usage (default command) |
| `serve` | Start MCP server on stdio |
| `install-script` | Install status line scripts |

## Development

```bash
# Build
just build

# Run tests
just test

# Install locally
just install

# Build for all platforms
just release
```

## License

BSD-3-Clause
