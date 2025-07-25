#!/usr/bin/env bash
set -euo pipefail

docker exec -i postgres \
  psql -U 33ccdb2917 -d ctfservice < $1

echo "🎉 migration.sql applied!"
