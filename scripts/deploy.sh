#!/bin/bash
set -e

# Configuration
SERVER_USER="infra"
SERVER_IP="88.198.170.206"
SERVER_PATH="/home/infra/savegress-platform"

echo "=== Deploying Savegress Platform ==="

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

# Ensure .env symlink exists in docker folder
echo ">>> Ensuring .env symlink..."
ssh "${SERVER_USER}@${SERVER_IP}" "ln -sf ${SERVER_PATH}/.env ${SERVER_PATH}/docker/.env 2>/dev/null || true"

# Deploy
echo ">>> Building and deploying..."
ssh "${SERVER_USER}@${SERVER_IP}" "cd ${SERVER_PATH}/docker && docker compose up -d --build"

# Check status
echo ">>> Checking deployment status..."
ssh "${SERVER_USER}@${SERVER_IP}" "cd ${SERVER_PATH}/docker && docker compose ps"

echo ""
echo "=== Deployment complete ==="
echo "Site: http://${SERVER_IP}"
echo ""
echo "Commands:"
echo "  make server-status  - Check status"
echo "  make server-logs    - View logs"
echo "  make server-ssh     - SSH to server"
