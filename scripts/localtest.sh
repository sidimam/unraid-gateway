#!/bin/sh
# Runs a throw-away gateway on 127.0.0.1:8585 against the real Unraid API for web UI / API tests.
# Usage: scripts/localtest.sh <workdir>   (the directory is created; keep it OUTSIDE the repo)
# The stored web UI key is copied from the NAS (ssh alias `nas`) so the browser can sign in with the
# "stored key" button — never type the key. Data root = <workdir>/data with a sample share.
set -e
W="${1:?workdir}"; mkdir -p "$W/config" "$W/data/testshare"
[ -f "$W/data/testshare/hello.txt" ] || echo hello > "$W/data/testshare/hello.txt"
[ -s "$W/config/webui.key" ] || { ssh nas 'cat /mnt/user/appdata/unraid-gateway/webui.key' > "$W/config/webui.key"; chmod 600 "$W/config/webui.key"; }
go build -ldflags "-X github.com/sidimam/unraid-gateway/internal/server.Version=$(git describe --tags --always)-local" -o "$W/gw" ./cmd/unraid-gateway
pkill -f "$W/gw" 2>/dev/null || true
cd "$W"
LISTEN_ADDR=127.0.0.1:8585 UNRAID_URL="${UNRAID_URL:-http://192.168.0.100}" DATA_ROOT="$W/data" INDEX_DB=off \
DEVICES_FILE="$W/config/devices.json" WEBUI_API_KEY_FILE="$W/config/webui.key" USER_AUTH=off DEVICE_REGISTRATION=on LOG_LEVEL=debug \
nohup "$W/gw" > "$W/gw.log" 2>&1 &
sleep 2; curl -s -m 5 http://127.0.0.1:8585/healthz; echo; echo "log: $W/gw.log  stop: pkill -f $W/gw"
