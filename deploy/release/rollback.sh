#!/usr/bin/env bash
# Emergency Rollback Runner (Restores previous digest state without schema destruction)
set -euo pipefail

PREVIOUS_STATE_FILE="/var/lib/deploy/knowledge-graph-platform.previous.env"

if [ ! -f "$PREVIOUS_STATE_FILE" ]; then
  echo "[FATAL] No previous deployment state found at ${PREVIOUS_STATE_FILE}. Cannot rollback safely."
  exit 1
fi

# Load previous digest
# shellcheck disable=SC1090
source "$PREVIOUS_STATE_FILE"

: "${PREVIOUS_APP_DIGEST:?PREVIOUS_APP_DIGEST missing in state file}"

echo "[INFO] Rolling back to APP_DIGEST=${PREVIOUS_APP_DIGEST}..."
APP_DIGEST="${PREVIOUS_APP_DIGEST}" docker compose -f deploy/docker/compose/docker-compose.single.yml up -d --wait --wait-timeout 180

echo "[INFO] Rollback completed successfully."
