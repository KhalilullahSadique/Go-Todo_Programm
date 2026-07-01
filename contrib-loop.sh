#!/usr/bin/env bash
# Commits + pushes to main on a loop until midnight (local time).
# Usage: ./contrib-loop.sh [seconds-between-commits]
set -euo pipefail

INTERVAL="${1:-60}"          # default: 60s between commits
BRANCH="main"
LOG=".contrib/log.txt"

mkdir -p "$(dirname "$LOG")"

# Midnight tonight, as an epoch timestamp.
END=$(date -d 'tomorrow 00:00:00' +%s)

echo "Looping every ${INTERVAL}s until $(date -d @"$END"). Ctrl-C to stop."

while [ "$(date +%s)" -lt "$END" ]; do
  echo "activity $(date --iso-8601=seconds)" >> "$LOG"
  git add "$LOG"
  git commit -q -m "chore: activity log $(date +%H:%M:%S)"
  if git push -q origin "$BRANCH"; then
    echo "pushed at $(date +%H:%M:%S)"
  else
    echo "push failed at $(date +%H:%M:%S) — will retry next loop"
  fi
  # Stop if the next sleep would cross midnight.
  [ "$(( $(date +%s) + INTERVAL ))" -ge "$END" ] && break
  sleep "$INTERVAL"
done

echo "Done at $(date +%H:%M:%S)."
