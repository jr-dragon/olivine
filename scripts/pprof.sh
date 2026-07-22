#!/usr/bin/env bash

set -Eeuo pipefail

project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output_dir="$project_root/scripts/out"
olivine_pid=""
benchmark_pid=""

cleanup() {
  local status=$?

  trap - EXIT INT TERM

  if [[ -n "$benchmark_pid" ]] && kill -0 "$benchmark_pid" 2>/dev/null; then
    kill "$benchmark_pid" 2>/dev/null || true
    wait "$benchmark_pid" 2>/dev/null || true
  fi

  if [[ -n "$olivine_pid" ]] && kill -0 "$olivine_pid" 2>/dev/null; then
    kill "$olivine_pid" 2>/dev/null || true
    wait "$olivine_pid" 2>/dev/null || true
  fi

  make clean
  exit "$status"
}

usage() {
  echo "Usage: $0 {profile|block|mutex|allocs|heap} [redis-benchmark flags...]" >&2
}

wait_for_pprof() {
  local attempt

  for attempt in {1..30}; do
    if curl --fail --silent --output /dev/null --max-time 1 \
      http://127.0.0.1:6060/debug/pprof/; then
      return 0
    fi
    sleep 1
  done

  echo "Olivine pprof server did not become ready on 127.0.0.1:6060." >&2
  return 1
}

if [[ $# -lt 1 ]]; then
  usage
  exit 2
fi

profile_type="$1"
shift

server_env=""
case "$profile_type" in
  profile|allocs|heap)
    ;;
  block)
    server_env="PPROF_BLOCK_RATE=1"
    ;;
  mutex)
    server_env="PPROF_MUTEX_FRAC=1"
    ;;
  *)
    usage
    exit 2
    ;;
esac

if ! command -v redis-benchmark >/dev/null 2>&1; then
  echo "redis-benchmark is required but was not found in PATH." >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required but was not found in PATH." >&2
  exit 1
fi

cd "$project_root"
trap cleanup EXIT INT TERM

make gen
make build

mkdir -p "$output_dir"
if [[ -n "$server_env" ]]; then
  env "$server_env" ./bin/olivine &
else
  ./bin/olivine &
fi
olivine_pid=$!

wait_for_pprof

datestr=$(date +%Y-%m-%d_%H:%M:%S)
profile_path="$output_dir/$datestr.$profile_type"

if [[ "$profile_type" == "profile" ]]; then
  redis-benchmark "$@" -p 16379 &
  benchmark_pid=$!

  curl --fail --show-error --output "$profile_path" \
    "http://localhost:6060/debug/pprof/$profile_type"
else
  redis-benchmark "$@" -p 16379

  curl --fail --show-error --output "$profile_path" \
    "http://localhost:6060/debug/pprof/$profile_type"
fi
