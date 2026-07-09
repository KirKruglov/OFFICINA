#!/usr/bin/env bash
# Claude Code status line script

input=$(cat)

model=$(echo "$input" | jq -r '.model.display_name // "Unknown model"')
effort=$(echo "$input" | jq -r '.effort.level // empty')
session=$(echo "$input" | jq -r '.session_name // empty')
used_pct=$(echo "$input" | jq -r '.context_window.used_percentage // empty')
used_tokens=$(echo "$input" | jq -r '.context_window.total_input_tokens // empty')
total_tokens=$(echo "$input" | jq -r '.context_window.context_window_size // empty')
cwd=$(echo "$input" | jq -r '.workspace.current_dir // .cwd // ""')
five_pct=$(echo "$input" | jq -r '.rate_limits.five_hour.used_percentage // empty')
five_resets=$(echo "$input" | jq -r '.rate_limits.five_hour.resets_at // empty')
week_pct=$(echo "$input" | jq -r '.rate_limits.seven_day.used_percentage // empty')
week_resets=$(echo "$input" | jq -r '.rate_limits.seven_day.resets_at // empty')

# Git branch (skip optional locks to avoid conflicts)
git_branch=""
if [ -n "$cwd" ] && [ -d "$cwd/.git" ] || git -C "$cwd" rev-parse --git-dir >/dev/null 2>&1; then
    git_branch=$(GIT_OPTIONAL_LOCKS=0 git -C "$cwd" symbolic-ref --short HEAD 2>/dev/null)
fi

# Build status line parts
parts=()

# Model name (cyan) + effort level (dim gray, if present)
if [ -n "$effort" ]; then
    parts+=("$(printf '\033[0;36m%s\033[0m \033[0;90m(%s)\033[0m' "$model" "$effort")")
else
    parts+=("$(printf '\033[0;36m%s\033[0m' "$model")")
fi

# Context: progress bar + percentage + absolute token count
if [ -n "$used_pct" ]; then
    used_int=$(printf '%.0f' "$used_pct")

    # Pick color based on usage level
    if [ "$used_int" -ge 80 ]; then
        color='\033[0;31m'   # red
    elif [ "$used_int" -ge 50 ]; then
        color='\033[0;33m'   # yellow
    else
        color='\033[0;32m'   # green
    fi

    # Build 10-char progress bar (█ filled, ░ empty)
    filled=$(( used_int * 10 / 100 ))
    [ "$filled" -gt 10 ] && filled=10
    empty=$(( 10 - filled ))
    bar=""
    for _ in $(seq 1 "$filled" 2>/dev/null); do bar="${bar}█"; done
    for _ in $(seq 1 "$empty"  2>/dev/null); do bar="${bar}░"; done

    # Format absolute token counts (convert to k for readability)
    if [ -n "$used_tokens" ] && [ -n "$total_tokens" ]; then
        used_k=$(echo "$used_tokens" | awk '{printf "%.0fk", $1/1000}')
        total_k=$(echo "$total_tokens" | awk '{printf "%.0fk", $1/1000}')
        ctx_str="${bar} ${used_int}% (${used_k}/${total_k})"
    else
        ctx_str="${bar} ${used_int}%"
    fi

    parts+=("$(printf "${color}%s\033[0m" "$ctx_str")")
fi

# 5-hour rate limit (magenta/red/yellow/green based on usage)
if [ -n "$five_pct" ]; then
    five_int=$(printf '%.0f' "$five_pct")

    if [ "$five_int" -ge 80 ]; then
        five_color='\033[0;31m'   # red
    elif [ "$five_int" -ge 50 ]; then
        five_color='\033[0;33m'   # yellow
    else
        five_color='\033[0;35m'   # magenta
    fi

    # Time until reset
    resets_str=""
    if [ -n "$five_resets" ]; then
        now=$(date +%s)
        diff=$(( five_resets - now ))
        if [ "$diff" -gt 0 ]; then
            mins=$(( diff / 60 ))
            hrs=$(( mins / 60 ))
            mins_rem=$(( mins % 60 ))
            if [ "$hrs" -gt 0 ]; then
                resets_str=" ~${hrs}h${mins_rem}m"
            else
                resets_str=" ~${mins}m"
            fi
        fi
    fi

    parts+=("$(printf "${five_color}5h:${five_int}%%${resets_str}\033[0m")")
fi

# Weekly rate limit (blue/red/yellow based on usage)
if [ -n "$week_pct" ]; then
    week_int=$(printf '%.0f' "$week_pct")

    if [ "$week_int" -ge 80 ]; then
        week_color='\033[0;31m'   # red
    elif [ "$week_int" -ge 50 ]; then
        week_color='\033[0;33m'   # yellow
    else
        week_color='\033[0;34m'   # blue
    fi

    # Time until reset
    week_resets_str=""
    if [ -n "$week_resets" ]; then
        now=$(date +%s)
        diff=$(( week_resets - now ))
        if [ "$diff" -gt 0 ]; then
            mins=$(( diff / 60 ))
            hrs=$(( mins / 60 ))
            mins_rem=$(( mins % 60 ))
            days=$(( hrs / 24 ))
            hrs_rem=$(( hrs % 24 ))
            if [ "$days" -gt 0 ]; then
                week_resets_str=" ~${days}d${hrs_rem}h"
            elif [ "$hrs" -gt 0 ]; then
                week_resets_str=" ~${hrs}h${mins_rem}m"
            else
                week_resets_str=" ~${mins}m"
            fi
        fi
    fi

    parts+=("$(printf "${week_color}wk:${week_int}%%${week_resets_str}\033[0m")")
fi

# Git branch (purple)
if [ -n "$git_branch" ]; then
    parts+=("$(printf '\033[0;37m\xef\xa0\xa1 %s\033[0m' "$git_branch")")
fi

# Join parts with separator
IFS='|'
printf '%s' "${parts[*]}"
