#!/bin/bash
set -e

if [ $# -lt 1 ]; then
	echo "usage: $0 <config.toml>" >&2
	exit 1
fi
config=$(realpath "$1")

repo_root=$(git rev-parse --show-toplevel)
pids=()

cleanup() {
	trap - INT TERM EXIT
	for pid in "${pids[@]}"; do
		kill "$pid" 2>/dev/null
	done
	wait 2>/dev/null || true
}
trap cleanup INT TERM EXIT

cd "$repo_root/backend"
go run -tags insecurecors cmd/eggplant/main.go run --verbosity debug "$config" &
pids+=($!)

cd "$repo_root/frontend"
corepack yarn serve --host 0.0.0.0 --clearScreen false &
pids+=($!)

while kill -0 "${pids[@]}" 2>/dev/null; do
	sleep 1
done
