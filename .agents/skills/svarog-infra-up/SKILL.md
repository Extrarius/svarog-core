---
name: svarog-infra-up
description: Bring up the full svarog-core local stack on WSL (Docker Desktop, compose migrate+app+mcp-http+LGTM) and run auth/MCP smoke checks. Use when the user asks to start the project, run make up, docker compose, verify localhost:8080 or :8000/mcp, or debug WSL Docker/npx path issues.
---

# svarog-infra-up

Start and smoke-test **svarog-core** infrastructure in this repository.

## Prerequisites

- **WSL2** with the repo at a Linux path (e.g. `/home/falih/svarog-core`), not only `\\wsl.localhost\...` for builds.
- **Docker Desktop** running on Windows (agent may start it via PowerShell if down).
- **Go** on PATH for local `make run`; **native Linux npx** via nvm (`~/.nvm`) — not Windows `/mnt/c/.../npx` for MCP/skills.

## Start Docker Desktop (WSL)

If `docker` is missing in WSL:

```bash
powershell.exe -Command "Start-Process 'C:\Program Files\Docker\Docker\Docker Desktop.exe'"
# wait until docker info works
DOCKER="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe"
"$DOCKER" info
```

## Bring up the stack

From repo root:

```bash
cd /home/falih/svarog-core
export PATH="$PWD/tools/apt-make/usr/bin:$PATH"   # if GNU make not installed
make up
# or: ./scripts/up.sh
```

This starts: **postgres**, **migrate** (one-shot), **app** (:8080, :9090), **mcp-http** (:8000/mcp), **otel-collector**, **loki**, **tempo**, **mimir**, **grafana** (:3000).

If **mimir** restarts, check `deploy/mimir/mimir.yaml` ruler paths (must not overlap).

## Smoke tests (HTTP)

```bash
# register
curl -s -X POST http://127.0.0.1:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"smoke@example.com","password":"secret123"}'

# login + cookie
curl -s -c /tmp/sv-cookies.txt -X POST http://127.0.0.1:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"smoke@example.com","password":"secret123"}'

# me (authenticated)
curl -s -b /tmp/sv-cookies.txt http://127.0.0.1:8080/v1/auth/me

# logout
curl -s -b /tmp/sv-cookies.txt -X POST http://127.0.0.1:8080/v1/auth/logout

# me after logout → expect unauthenticated
curl -s -b /tmp/sv-cookies.txt http://127.0.0.1:8080/v1/auth/me

# MCP HTTP endpoint reachable
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8000/mcp
```

## Local run without Docker app

```bash
make up          # infra only, or full stack
make migrate     # if migrate job not used
make run         # terminal 1 — app
make run-mcp-http  # optional second terminal
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `make: not found` | `export PATH="$PWD/tools/apt-make/usr/bin:$PATH"` or `sudo apt install make` |
| `docker` not found in WSL | Use Docker Desktop + `docker.exe` path or enable WSL integration |
| Login HTTP 500, sessions in DB | Fixed: scan `inet` as text in `internal/adapters/repo/sessions.go` |
| context7 / npx red in Cursor | Use remote URL in `.cursor/mcp.json`, not Windows npx |

## References

- [`README.md`](../../../README.md), [`doc/MCP.md`](../../../doc/MCP.md)
- Compose: [`deploy/docker-compose.yml`](../../../deploy/docker-compose.yml)
