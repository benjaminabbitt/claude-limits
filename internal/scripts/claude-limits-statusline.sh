#!/bin/bash
# Claude Code status line showing Claude.ai usage limits

# Read JSON from stdin (Claude Code status line protocol)
INPUT=$(cat)

# Parse context window data from stdin
CONTEXT_SIZE=$(echo "$INPUT" | jq -r '.context_window.context_window_size // empty' 2>/dev/null)
CURRENT_USAGE=$(echo "$INPUT" | jq -r '.context_window.current_usage // empty' 2>/dev/null)

# Calculate context utilization from stdin if available
if [[ -n "$CURRENT_USAGE" && "$CURRENT_USAGE" != "null" ]]; then
    INPUT_TOKENS=$(echo "$INPUT" | jq -r '.context_window.current_usage.input_tokens // 0' 2>/dev/null)
    OUTPUT_TOKENS=$(echo "$INPUT" | jq -r '.context_window.current_usage.output_tokens // 0' 2>/dev/null)
    CACHE_CREATE=$(echo "$INPUT" | jq -r '.context_window.current_usage.cache_creation_input_tokens // 0' 2>/dev/null)
    CACHE_READ=$(echo "$INPUT" | jq -r '.context_window.current_usage.cache_read_input_tokens // 0' 2>/dev/null)
    # Ensure numeric values (default to 0 if empty or non-numeric)
    INPUT_TOKENS=${INPUT_TOKENS:-0}
    OUTPUT_TOKENS=${OUTPUT_TOKENS:-0}
    CACHE_CREATE=${CACHE_CREATE:-0}
    CACHE_READ=${CACHE_READ:-0}
    [[ "$INPUT_TOKENS" =~ ^[0-9]+$ ]] || INPUT_TOKENS=0
    [[ "$OUTPUT_TOKENS" =~ ^[0-9]+$ ]] || OUTPUT_TOKENS=0
    [[ "$CACHE_CREATE" =~ ^[0-9]+$ ]] || CACHE_CREATE=0
    [[ "$CACHE_READ" =~ ^[0-9]+$ ]] || CACHE_READ=0
    CURRENT_TOKENS=$((INPUT_TOKENS + OUTPUT_TOKENS + CACHE_CREATE + CACHE_READ))
    # Only calculate if CONTEXT_SIZE is a positive number
    if [[ "$CONTEXT_SIZE" =~ ^[0-9]+$ && "$CONTEXT_SIZE" -gt 0 ]]; then
        STDIN_CONTEXT=$((CURRENT_TOKENS * 100 / CONTEXT_SIZE))
    fi
fi

# ANSI color codes
RED='\033[31m'
YELLOW='\033[33m'
GREEN='\033[32m'
RESET='\033[0m'

# Time format (strftime syntax, default to 12-hour format)
# Use "+%H:%M" for 24-hour, "+%I:%M %p" for 12-hour with AM/PM
TIME_FORMAT="${CLAUDE_LIMITS_TIME_FORMAT:-+%I:%M %p}"
# Date+time format for weekly reset (includes day)
DATETIME_FORMAT="${CLAUDE_LIMITS_DATETIME_FORMAT:-+%a %I:%M %p}"

# Colorize a percentage value based on thresholds
colorize() {
    local value="$1"
    if [[ "$value" == "?" ]]; then
        echo "$value"
        return
    fi
    # Extract numeric part (remove any decimal)
    local num="${value%.*}"
    if [[ "$num" -ge 95 ]]; then
        echo -e "${RED}${value}${RESET}"
    elif [[ "$num" -ge 80 ]]; then
        echo -e "${YELLOW}${value}${RESET}"
    else
        echo -e "${GREEN}${value}${RESET}"
    fi
}

# Format an ISO timestamp to local time
format_time() {
    local iso_time="$1"
    local format="$2"

    if [[ -z "$iso_time" || "$iso_time" == "?" ]]; then
        echo "?"
        return
    fi

    # Try GNU date first (Linux), then BSD date (macOS)
    if date --version >/dev/null 2>&1; then
        # GNU date
        date -d "$iso_time" "$format" 2>/dev/null || echo "?"
    else
        # BSD date (macOS) - convert ISO 8601 to epoch then format
        local epoch
        epoch=$(date -j -f "%Y-%m-%dT%H:%M:%S%z" "${iso_time//Z/+0000}" '+%s' 2>/dev/null) ||
        epoch=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$iso_time" '+%s' 2>/dev/null)
        if [[ -n "$epoch" ]]; then
            date -j -f '%s' "$epoch" "$format" 2>/dev/null || echo "?"
        else
            echo "?"
        fi
    fi
}

# Find claude-limits binary
CLAUDE_LIMITS="${CLAUDE_LIMITS_PATH:-$(command -v claude-limits 2>/dev/null)}"
if [[ -z "$CLAUDE_LIMITS" ]]; then
    echo "claude-limits: not found"
    exit 1
fi

# Get utilization values and reset times (using specific queries to avoid ambiguity)
FIVE_HOUR=$($CLAUDE_LIMITS five_hour_utilization 2>/dev/null)
WEEKLY=$($CLAUDE_LIMITS seven_day_utilization 2>/dev/null)
# Use context from stdin if available, otherwise try claude-limits
CONTEXT="${STDIN_CONTEXT:-$($CLAUDE_LIMITS context_utilization 2>/dev/null)}"
FIVE_HOUR_RESET=$($CLAUDE_LIMITS five_hour_reset 2>/dev/null)
WEEKLY_RESET=$($CLAUDE_LIMITS seven_day_reset 2>/dev/null)

# Default to "?" if not available
FIVE_HOUR=${FIVE_HOUR:-"?"}
WEEKLY=${WEEKLY:-"?"}
CONTEXT=${CONTEXT:-"?"}

# Format reset times
FIVE_HOUR_RESET_LOCAL=$(format_time "$FIVE_HOUR_RESET" "$TIME_FORMAT")
WEEKLY_RESET_LOCAL=$(format_time "$WEEKLY_RESET" "$DATETIME_FORMAT")

# Colorize values
FIVE_HOUR_C=$(colorize "$FIVE_HOUR")
WEEKLY_C=$(colorize "$WEEKLY")
CONTEXT_C=$(colorize "$CONTEXT")

# Output the status line
echo -e "5h: ${FIVE_HOUR_C}% @ ${FIVE_HOUR_RESET_LOCAL} | wk: ${WEEKLY_C}% @ ${WEEKLY_RESET_LOCAL} | ctx: ${CONTEXT_C}%"
