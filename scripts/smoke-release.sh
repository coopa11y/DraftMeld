#!/usr/bin/env sh
set -eu

repository_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
version="$(tr -d '\r\n' < "$repository_root/VERSION")"
archive="${1:-$repository_root/outputs/release/draftmeld_${version}_linux_amd64.tar.gz}"
temporary_directory="$(mktemp -d)"
port="${DRAFTMELD_SMOKE_PORT:-18080}"
process_id=""

cleanup() {
  if [ -n "$process_id" ]; then
    kill "$process_id" 2>/dev/null || true
    wait "$process_id" 2>/dev/null || true
  fi
  rm -rf -- "$temporary_directory"
}
trap cleanup EXIT INT TERM

tar -C "$temporary_directory" -xzf "$archive"
binary="$temporary_directory/draftmeld_${version}_linux_amd64/draftmeld"
test -x "$binary"

DRAFTMELD_ADDRESS="127.0.0.1:$port" DRAFTMELD_DATA_DIR="$temporary_directory/data" "$binary" > "$temporary_directory/server.log" 2>&1 &
process_id="$!"

attempt=0
while [ "$attempt" -lt 30 ]; do
  if response="$(curl --fail --silent "http://127.0.0.1:$port/api/v1/health")"; then
    HEALTH_RESPONSE="$response" EXPECTED_VERSION="$version" node -e '
      const health = JSON.parse(process.env.HEALTH_RESPONSE);
      if (health.status !== "ok" || health.version !== process.env.EXPECTED_VERSION) {
        throw new Error(`Unexpected health response: ${JSON.stringify(health)}`);
      }
    '
    echo "DraftMeld $version release smoke test passed."
    exit 0
  fi
  if ! kill -0 "$process_id" 2>/dev/null; then
    cat "$temporary_directory/server.log"
    exit 1
  fi
  attempt=$((attempt + 1))
  sleep 1
done

cat "$temporary_directory/server.log"
echo "DraftMeld did not become healthy within 30 seconds." >&2
exit 1
