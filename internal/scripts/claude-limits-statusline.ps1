# Claude Code status line showing Claude.ai usage limits (Windows)

# Read JSON from stdin (Claude Code status line protocol)
$InputText = @($input) -join "`n"
$StdinContext = $null

# Parse context window data from stdin
if ($InputText) {
    try {
        $InputData = $InputText | ConvertFrom-Json -ErrorAction Stop
        $ContextSize = $InputData.context_window.context_window_size
        $CurrentUsage = $InputData.context_window.current_usage

        if ($CurrentUsage -and $ContextSize -gt 0) {
            # PS 5.1 compatible null handling (no ?? operator)
            $InputTokens = if ($null -ne $CurrentUsage.input_tokens) { [int]$CurrentUsage.input_tokens } else { 0 }
            $CacheCreate = if ($null -ne $CurrentUsage.cache_creation_input_tokens) { [int]$CurrentUsage.cache_creation_input_tokens } else { 0 }
            $CacheRead = if ($null -ne $CurrentUsage.cache_read_input_tokens) { [int]$CurrentUsage.cache_read_input_tokens } else { 0 }
            $CurrentTokens = $InputTokens + $CacheCreate + $CacheRead
            $StdinContext = [math]::Floor($CurrentTokens * 100 / $ContextSize)
        }
    } catch {
        # Ignore JSON parse errors
    }
}

# ANSI color codes
$RED = "`e[31m"
$YELLOW = "`e[33m"
$GREEN = "`e[32m"
$RESET = "`e[0m"

# Colorize a percentage value based on thresholds
function Colorize {
    param([string]$Value)
    if ($Value -eq "?" -or -not $Value) {
        return "?"
    }
    $num = [int][math]::Floor([double]$Value)
    if ($num -ge 95) {
        return "${RED}${Value}${RESET}"
    } elseif ($num -ge 80) {
        return "${YELLOW}${Value}${RESET}"
    } else {
        return "${GREEN}${Value}${RESET}"
    }
}

# Find claude-limits binary
$CLAUDE_LIMITS = if ($env:CLAUDE_LIMITS_PATH) {
    $env:CLAUDE_LIMITS_PATH
} else {
    Get-Command claude-limits -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source
}

if (-not $CLAUDE_LIMITS) {
    Write-Output "claude-limits: not found"
    exit 1
}

# Get utilization values and reset time (using specific queries to avoid ambiguity)
try {
    $FIVE_HOUR = & $CLAUDE_LIMITS five_hour_utilization 2>$null
} catch {
    $FIVE_HOUR = $null
}

try {
    $WEEKLY = & $CLAUDE_LIMITS seven_day_utilization 2>$null
} catch {
    $WEEKLY = $null
}

# Use context from stdin if available, otherwise try claude-limits
$CONTEXT = $StdinContext
if (-not $CONTEXT) {
    try {
        $CONTEXT = & $CLAUDE_LIMITS context_utilization 2>$null
    } catch {
        $CONTEXT = $null
    }
}

try {
    $RESET_TIME = & $CLAUDE_LIMITS five_hour_reset 2>$null
} catch {
    $RESET_TIME = $null
}

# Default to "?" if not available
if (-not $FIVE_HOUR) { $FIVE_HOUR = "?" }
if (-not $WEEKLY) { $WEEKLY = "?" }
if (-not $CONTEXT) { $CONTEXT = "?" }

# Parse and localize reset time
$RESET_LOCAL = "?"
if ($RESET_TIME -and $RESET_TIME -ne "?") {
    try {
        $dt = [DateTime]::Parse($RESET_TIME)
        $RESET_LOCAL = $dt.ToLocalTime().ToString("HH:mm")
    } catch {
        $RESET_LOCAL = "?"
    }
}

# Colorize values
$FIVE_HOUR_C = Colorize $FIVE_HOUR
$WEEKLY_C = Colorize $WEEKLY
$CONTEXT_C = Colorize $CONTEXT

# Output the status line
Write-Output "5h: ${FIVE_HOUR_C}% @ ${RESET_LOCAL} | wk: ${WEEKLY_C}% | ctx: ${CONTEXT_C}%"
