#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATABASE_URL="${DATABASE_URL:-postgres://govbenefits:govbenefits@localhost:5432/govbenefits?sslmode=disable}"

echo "Resetting demo data..."
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$SCRIPT_DIR/reset_demo_data.sql"
echo "Loading curated demo dataset..."
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$SCRIPT_DIR/seed_demo_data.sql"
echo "Demo dataset ready."
