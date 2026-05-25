#!/bin/bash
set -e

public=0
args=()
for arg in "$@"; do
	if [ "$arg" = "--public" ]; then
		public=1
	else
		args+=("$arg")
	fi
done
set -- "${args[@]}"

if [ $# -lt 1 ]; then
	echo "usage: $0 [--public] <config.toml>" >&2
	exit 1
fi
config=$(realpath "$1")

if [ "$public" = "1" ]; then
	target=serve-public
else
	target=serve
fi

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

make -C "$repo_root/backend" run CONFIG="$config" &
pids+=($!)

make -C "$repo_root/frontend" "$target" &
pids+=($!)

while kill -0 "${pids[@]}" 2>/dev/null; do
	sleep 1
done
