#!/usr/bin/env bash

# ShelfScan Restore Helper Script
# This script restores the SQLite database, uploads directory, and Qdrant storage volume from a backup archive.

set -euo pipefail

BACKUP_DIR="./backups"
TEMP_RESTORE_DIR="./tmp_restore"

# Check if a backup file is provided
if [ $# -eq 0 ]; then
  echo "=== ShelfScan Backup Restore Utility ==="
  echo "Available backups in $BACKUP_DIR:"
  if [ -d "$BACKUP_DIR" ] && [ "$(ls -A "$BACKUP_DIR" 2>/dev/null)" ]; then
    ls -lh "$BACKUP_DIR"/shelfscan-backup-*.tar.gz 2>/dev/null || echo "No standard ShelfScan backup files found."
  else
    echo "No backup directory found or directory is empty."
  fi
  echo ""
  echo "Usage: $0 <path-to-backup-archive.tar.gz>"
  exit 1
fi

BACKUP_FILE="$1"

# Verify backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
  echo "Error: Backup file '$BACKUP_FILE' does not exist."
  exit 1
fi

echo "=== ShelfScan Restore ==="
echo "Backup File: $BACKUP_FILE"
echo "WARNING: This will overwrite your active SQLite database, uploaded files, and Qdrant vector database."
read -p "Are you sure you want to proceed? (y/N) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Restore cancelled."
  exit 0
fi

# Detect Docker Compose Project Name
PROJECT_NAME=$(docker compose config --format json | jq -r '.name' 2>/dev/null || echo "shelfscan")
QDRANT_VOLUME="${PROJECT_NAME}_qdrant_storage"

echo "Using Qdrant volume: $QDRANT_VOLUME"

# 1. Stop containers
echo "Stopping api and qdrant services..."
docker compose stop api qdrant

# Create temp restore staging directory
mkdir -p "$TEMP_RESTORE_DIR"

# Cleanup trap to ensure staging is deleted on exit/error
trap 'rm -rf "$TEMP_RESTORE_DIR"' EXIT

# 2. Extract backup
echo "Extracting backup archive..."
tar -xzf "$BACKUP_FILE" -C "$TEMP_RESTORE_DIR"

# 3. Restore SQLite database
if [ -d "$TEMP_RESTORE_DIR/backup/db" ]; then
  echo "Restoring SQLite database..."
  mkdir -p ./server/db
  rm -f ./server/db/shelfscan.db*
  cp -rf "$TEMP_RESTORE_DIR/backup/db/." ./server/db/
else
  echo "Warning: No SQLite database folder found in backup."
fi

# 4. Restore Uploads
if [ -d "$TEMP_RESTORE_DIR/backup/uploads" ]; then
  echo "Restoring uploaded media..."
  mkdir -p ./server/uploads
  cp -rf "$TEMP_RESTORE_DIR/backup/uploads/." ./server/uploads/
else
  echo "Warning: No uploads folder found in backup."
fi

# 5. Restore Qdrant Volume
if [ -d "$TEMP_RESTORE_DIR/backup/qdrant" ]; then
  echo "Restoring Qdrant storage volume..."
  # Use a temporary alpine container to copy files directly into the named volume
  docker run --rm \
    -v "$QDRANT_VOLUME":/qdrant/storage \
    -v "$(pwd)/$TEMP_RESTORE_DIR/backup/qdrant":/restore \
    alpine sh -c "rm -rf /qdrant/storage/* && cp -a /restore/. /qdrant/storage/"
else
  echo "Warning: No Qdrant volume data found in backup."
fi

# 6. Restart containers
echo "Restarting services..."
docker compose start qdrant api

echo "=== Restore Completed Successfully! ==="
