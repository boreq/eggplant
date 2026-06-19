#!/bin/bash
# Runs two full Eggplant instances side by side so you can test the
# instance pairing / join-remotes flow locally.
#
# Each instance gets its own data + cache directory (so they have separate
# databases and users) but shares the music directory. Both the backend and
# the frontend dev server are started for each instance.
#
#   Instance A: backend http://127.0.0.1:8118  frontend http://127.0.0.1:5173
#   Instance B: backend http://127.0.0.1:8119  frontend http://127.0.0.1:5174
#
# Note: the auth-token cookie is scoped by host, not port, so both frontends
# share a cookie jar. Open instance B in a private/incognito window (or a
# second browser profile) to keep the two logins separate.
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
frontend_a_port=5173
frontend_b_port=5174

if [ ! -d "$music_dir" ]; then
	echo "music directory not found: $music_dir" >&2
	exit 1
fi

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
	corepack yarn serve --clearScreen false --port "$frontend_a_port" ) &
pids+=($!)
( cd "$repo_root/frontend" && \
	VITE_DEV_BACKEND="http://127.0.0.1:$backend_b_port" \
	corepack yarn serve --clearScreen false --port "$frontend_b_port" ) &
pids+=($!)

cat <<EOF

================ eggplant pairing dev ================
 Instance A  frontend http://127.0.0.1:$frontend_a_port  backend 127.0.0.1:$backend_a_port
 Instance B  frontend http://127.0.0.1:$frontend_b_port  backend 127.0.0.1:$backend_b_port

 Open instance B in an incognito window (shared cookie jar otherwise).

 Pair by pointing one instance at the other's BACKEND
 address, e.g. http://127.0.0.1:$backend_b_port
 Ctrl-C to stop both.
======================================================

EOF

while kill -0 "${pids[@]}" 2>/dev/null; do
	sleep 1
done
