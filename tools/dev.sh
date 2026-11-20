#!/bin/bash
set -e

repo_root=$(git rev-parse --show-toplevel)

music_dir="$repo_root/_dev/music"
state_root="$repo_root/_dev/data"

backend_a_port=8118
backend_b_port=8119
frontend_a_host=127.0.0.1
frontend_b_host=127.0.0.2
frontend_a_port=5173
frontend_b_port=5174

mkdir -p "$music_dir"

check_port_free() {
	local host=$1
	local port=$2
	if ss -tln "sport = :$port" 2>/dev/null | grep -q "$host:$port"; then
		echo "$host:$port is already in use, is another tools/dev.sh running?" >&2
		echo "stop it with: pkill -f tools/dev.sh; pkill -f 'air -c .air.toml'; pkill -f 'eggplant run'" >&2
		exit 1
	fi
}

check_port_free 127.0.0.1 "$backend_a_port"
check_port_free 127.0.0.1 "$backend_b_port"
check_port_free "$frontend_a_host" "$frontend_a_port"
check_port_free "$frontend_b_host" "$frontend_b_port"

write_config() {
	local dir=$1
	local port=$2
	mkdir -p "$dir/data" "$dir/cache"
	if [ -f "$dir/config.toml" ]; then
		return
	fi
	cat > "$dir/config.toml" <<EOF
serve_address = "127.0.0.1:$port"
music_directory = "$music_dir"
data_directory = "$dir/data"
cache_directory = "$dir/cache"
EOF
}

write_config "$state_root/a" "$backend_a_port"
write_config "$state_root/b" "$backend_b_port"

make -C "$repo_root/frontend" install

pids=()

cleanup() {
	trap - INT TERM EXIT
	for pid in "${pids[@]}"; do
		kill "$pid" 2>/dev/null
	done
	wait 2>/dev/null || true
}
trap cleanup INT TERM EXIT

run_backend() {
	local name=$1
	local build_dir="_build/air/$name"
	( cd "$repo_root/backend" && \
		air -c .air.toml \
			-tmp_dir "$build_dir" \
			-build.cmd "go build -o ./$build_dir/eggplant ./cmd/eggplant" \
			-build.entrypoint "./$build_dir/eggplant" \
			-build.args_bin "run,--verbosity,debug,$state_root/$name/config.toml" ) &
	pids+=($!)
}

run_frontend() {
	local host=$1
	local port=$2
	local backend_port=$3
	( cd "$repo_root/frontend" && \
		VITE_DEV_BACKEND="http://127.0.0.1:$backend_port" \
		corepack yarn serve --clearScreen false --strictPort \
			--host "$host" --port "$port" ) &
	pids+=($!)
}

run_backend a
run_backend b
run_frontend "$frontend_a_host" "$frontend_a_port" "$backend_a_port"
run_frontend "$frontend_b_host" "$frontend_b_port" "$backend_b_port"

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
