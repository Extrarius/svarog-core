#!/usr/bin/env bash
# Docker Compose helper for WSL2 + Docker Desktop (when `docker` is not on PATH).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT}/deploy/docker-compose.yml"

if command -v docker >/dev/null 2>&1; then
  DOCKER=(docker)
elif [[ -x "/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe" ]]; then
  DOCKER=("/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe")
else
  echo "docker not found: install Docker Desktop and enable WSL integration" >&2
  exit 1
fi

exec "${DOCKER[@]}" compose -f "${COMPOSE_FILE}" "$@"
