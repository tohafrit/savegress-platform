#!/bin/bash
set -e

# Configuration
SERVER_USER="infra"
SERVER_IP="88.198.170.206"
SERVER_PATH="/home/infra/savegress-platform"

echo "=== Quick Deploy (sync only, no rebuild) ==="

# Sync files
echo ">>> Syncing files to server..."
rsync -avz --progress \
  --exclude 'node_modules' \
  --exclude '.next' \
  --exclude 'dist' \
  --exclude '.git' \
  --exclude '*.log' \
  --exclude 'site.md' \
  --exclude '.env' \
  --exclude '.claude' \
  "$(dirname "$0")/../" \
  "${SERVER_USER}@${SERVER_IP}:${SERVER_PATH}/"

# Restart containers
echo ">>> Restarting containers..."
ssh "${SERVER_USER}@${SERVER_IP}" "cd ${SERVER_PATH}/docker && docker compose restart"

echo "=== Quick deploy complete ==="
