#!/usr/bin/env bash
set -euo pipefail

# Navigate to project root (one level up from scripts/)
cd "$(dirname "$0")/.."

echo "Regenerating OpenAPI docs..."
swag init -g main.go --output docs
echo "docs/ regenerated successfully."
