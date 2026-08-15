#!/bin/bash
# Runs two full Eggplant instances side by side so you can test the
# instance pairing / join-remotes flow locally.
#
# Each instance gets its own data + cache directory (so they have separate
# databases and users) but shares the music directory. Both the backend and
# the frontend dev server are started for each instance.
#
#   Instance A: backend http://127.0.0.1:8118  frontend http://127.0.0.1:5173
#   Instance B: backend http://127.0.0.1:8119  frontend http://127.0.0.2:5174
#
# Note: the auth-token cookie is scoped by host, not port, so the two frontends
# are bound to different loopback addresses (127.0.0.1 and 127.0.0.2). That
# gives them separate cookie jars, so both instances can be logged in at the
# same time in the same browser window.
#
# When pairing, point one instance at the other's BACKEND address
# (e.g. http://127.0.0.1:8119).
#
# Each run starts with fresh databases: the data directories are wiped so the
# two instances always come up empty and unpaired. The caches are kept so we
# do not re-transcode the whole music library every time.
#
# Usage:
#   ./tools/dev-pair.sh   # start both instances with fresh databases
set -e

repo_root=$(git rev-parse --show-toplevel)

music_dir="$repo_root/_misc/music"
state_root="$repo_root/_misc/pair"

backend_a_port=8118
backend_b_port=8119
# Different hosts, not just different ports: cookies ignore the port, so both
# frontends would otherwise share a single auth-token cookie.
frontend_a_host=127.0.0.1
frontend_b_host=127.0.0.2
frontend_a_port=5173
frontend_b_port=5174

if [ ! -d "$music_dir" ]; then
	echo "music directory not found: $music_dir" >&2
	exit 1
fi

# Bail out if a previous run is still around. Otherwise the old backends keep
# serving the databases we are about to wipe (they hold the files open) and the
# frontends silently move to other ports, so you end up talking to the stale
# instances without noticing.
check_port_free() {
	local host=$1
	local port=$2
	if ss -tln "sport = :$port" 2>/dev/null | grep -q "$host:$port"; then
		echo "$host:$port is already in use, is another dev-pair.sh running?" >&2
		echo "stop it with: pkill -f dev-pair.sh; pkill -f 'eggplant run'" >&2
		exit 1
	fi
}

check_port_free 127.0.0.1 "$backend_a_port"
check_port_free 127.0.0.1 "$backend_b_port"
check_port_free "$frontend_a_host" "$frontend_a_port"
check_port_free "$frontend_b_host" "$frontend_b_port"

# write_config <instance_dir> <backend_port>
write_config() {
	local dir=$1
	local port=$2
	# Wipe the database so each run starts fresh; keep the cache.
	rm -rf "$dir/data"
	mkdir -p "$dir/data" "$dir/cache"
	cat > "$dir/config.toml" <<EOF
serve_address = "127.0.0.1:$port"
music_directory = "$music_dir"
data_directory = "$dir/data"
cache_directory = "$dir/cache"
EOF
}

write_config "$state_root/a" "$backend_a_port"
write_config "$state_root/b" "$backend_b_port"

pids=()

cleanup() {
	trap - INT TERM EXIT
	for pid in "${pids[@]}"; do
		kill "$pid" 2>/dev/null
	done
	wait 2>/dev/null || true
}
trap cleanup INT TERM EXIT

# backend instances
make -C "$repo_root/backend" run CONFIG="$state_root/a/config.toml" &
pids+=($!)
make -C "$repo_root/backend" run CONFIG="$state_root/b/config.toml" &
pids+=($!)

# frontend dev servers, each proxying /api to its own backend
( cd "$repo_root/frontend" && \
	VITE_DEV_BACKEND="http://127.0.0.1:$backend_a_port" \
	corepack yarn serve --clearScreen false --strictPort \
		--host "$frontend_a_host" --port "$frontend_a_port" ) &
pids+=($!)
( cd "$repo_root/frontend" && \
	VITE_DEV_BACKEND="http://127.0.0.1:$backend_b_port" \
	corepack yarn serve --clearScreen false --strictPort \
		--host "$frontend_b_host" --port "$frontend_b_port" ) &
pids+=($!)

cat <<EOF

================ eggplant pairing dev ================
 Instance A  frontend http://$frontend_a_host:$frontend_a_port  backend 127.0.0.1:$backend_a_port
 Instance B  frontend http://$frontend_b_host:$frontend_b_port  backend 127.0.0.1:$backend_b_port

 The frontends use different hosts, so they have separate cookie
 jars - no incognito window needed.

 Pair by pointing one instance at the other's BACKEND
 address, e.g. http://127.0.0.1:$backend_b_port
 Ctrl-C to stop both.
======================================================

EOF

while kill -0 "${pids[@]}" 2>/dev/null; do
	sleep 1
done
