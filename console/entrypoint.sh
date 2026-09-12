#!/bin/sh
# Runs as root just long enough to make Unraid's notification spool writable for the gateway
# (host /tmp/notifications is root-owned 755; the gateway runs as nobody:users = 99:100, like files
# created over SMB). Everything else runs unprivileged, as before.
dir="${UNRAID_NOTIFY_DIR:-/unraid-notifications}"
if [ -d "$dir" ]; then
  mkdir -p "$dir/unread" 2>/dev/null
  chmod 1777 "$dir/unread" 2>/dev/null || true
fi
exec su-exec 99:100 /usr/local/bin/unraid-gateway "$@"
