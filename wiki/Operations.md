# Operations

## Health

`GET /healthz` → `{"status":"ok","version":"…"}`. The image defines a Docker `HEALTHCHECK` on it; the Unraid template adds `--health-cmd` so the Docker tab shows *healthy*.

## Logs

One JSON line per request on stdout (`LOG_REQUESTS=true`):

```json
{"time":"…","level":"INFO","msg":"request","method":"PUT","path":"/api/v1/fs/content","status":201,"ms":42,"ip":"203.0.113.7"}
```

Other lines: `unraid-gateway listening` (startup, with config summary), `login`, `login failed`, `login locked`, `unraid validation error`, `fs error`, `upload aborted`. Docker tab → icon → **Logs**.

## Updating

The template uses `latest`; Unraid's *Check for Updates* compares digests with GHCR and *Update* recreates the container from the template. *CA Auto Update Applications* can automate it. Sessions are lost on update; clients log in again transparently.

## Backups

Nothing to back up: the gateway holds no state besides in-memory sessions and in-flight upload temp files (`.name.gwpart*` next to the destination, removed on commit/abort/expiry).

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| Container exits immediately | `UNRAID_URL` missing or without scheme; a mapped host path does not exist. Read the first log lines. |
| `invalid api key` for a good key | `UNRAID_URL` wrong (must be the LAN IP, `http://` unless the WebGUI forces HTTPS), Unraid API stopped (`unraid-api status`), or key deleted. Log shows `unraid validation error`. |
| `401 invalid unraid username or password` | Samba rejected the user (wrong password, or unknown user mapped to guest and failing the share probe). Reset the password in the Unraid WebGUI → Users. |
| `401 this Unraid user has no access to any share mounted in the gateway` | Unraid's share security grants the user nothing on the mapped shares: add the user to the share's read/write list, or map another share. |
| `502 the Unraid share configuration is not readable by the gateway` | Mount `/etc/samba/smb-shares.conf` → `/unraid-shares/smb-shares.conf:ro` (template default) or set `USER_AUTH=off`. |
| `429 too many failed attempts` | lockout; wait `LOGIN_LOCKOUT` or restart. Behind a proxy without `TRUST_PROXY` everyone shares one IP. |
| `413` on uploads from the web UI | proxy body limit (Cloudflare 100 MB, nginx default 1 MB). Raise it or use the app. |
| Writes fail with `share is mounted read-only` | the mount has `:ro`. |
| `404 share not found` on write | the first path component is not a mounted share (typo, or share not mapped). |
| Change feed always truncated | very large share: normal; clients page with `after`. Raise `CHANGES_DEADLINE` if you prefer fewer, longer pages. |
| Dates off by hours | display only: API times are UTC; set `TZ` for log timestamps. |
