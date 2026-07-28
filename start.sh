#!/usr/bin/env bash
set -Eeuo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
mode="all"

case "${1:-}" in
	"") ;;
	--infra-only) mode="infra-only" ;;
	--no-infra) mode="no-infra" ;;
	--help|-h)
		echo "Usage: ./start.sh [--infra-only|--no-infra]"
		echo "  --infra-only  Start only Postgres and Redis with Docker"
		echo "  --no-infra    Run the apps against already-running local services"
		exit 0
		;;
	*)
		echo "Unknown option: $1 (try --help)" >&2
		exit 2
		;;
esac

if [[ "$mode" != "no-infra" ]]; then
	if ! command -v docker >/dev/null 2>&1; then
		echo "Missing required command: docker (or use --no-infra)." >&2
		exit 1
	fi
	if ! docker compose version >/dev/null 2>&1; then
		echo "Docker Compose v2 is required." >&2
		exit 1
	fi
	docker compose --project-directory "$repo_dir" up -d --wait postgres redis
fi

if [[ "$mode" == "infra-only" ]]; then
	echo "Postgres and Redis are ready."
	exit 0
fi

for command_name in go bun; do
	if ! command -v "$command_name" >/dev/null 2>&1; then
		echo "Missing required command: $command_name" >&2
		exit 1
	fi
done

if [[ ! -f "$repo_dir/api/.env" ]]; then
	cp "$repo_dir/api/.env.example" "$repo_dir/api/.env"
	echo "Created api/.env from the local development defaults."
fi
if [[ ! -f "$repo_dir/frontend/.env.local" ]]; then
	cp "$repo_dir/frontend/.env.example" "$repo_dir/frontend/.env.local"
	echo "Created frontend/.env.local from the local development defaults."
fi

if [[ ! -d "$repo_dir/frontend/node_modules" ]]; then
	(cd "$repo_dir/frontend" && bun install)
fi

cleanup() {
	trap - INT TERM EXIT
	kill "${api_pid:-}" "${frontend_pid:-}" 2>/dev/null || true
	wait "${api_pid:-}" "${frontend_pid:-}" 2>/dev/null || true
}
trap cleanup INT TERM EXIT

echo "Applying database migrations"
(cd "$repo_dir/api" && go run ./cmd/migrate)

echo "Starting API on http://localhost:8080"
if command -v air >/dev/null 2>&1; then
	(cd "$repo_dir/api" && air) &
else
	echo "Tip: install air for backend hot reload; falling back to go run."
	(cd "$repo_dir/api" && go run ./cmd/server) &
fi
api_pid=$!

echo "Starting frontend on http://localhost:3000"
(cd "$repo_dir/frontend" && bun run dev) &
frontend_pid=$!

wait -n "$api_pid" "$frontend_pid"
