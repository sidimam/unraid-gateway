# Configuration reference

All configuration is by environment variable. Booleans accept `true`/`false`; durations use Go syntax (`30s`, `15m`, `12h`, `168h`).

| Variable | Default | Description |
|---|---|---|
| `UNRAID_URL` | **required** | Base URL of the Unraid WebGUI on the LAN, e.g. `http://192.168.0.100`. Used to validate API keys (`POST /graphql`) and as the GraphQL proxy target. Use the IP, not `localhost` (the container has its own network namespace). |
| `UNRAID_INSECURE_TLS` | `false` | Skip certificate verification when `UNRAID_URL` is `https://` with a self-signed certificate. |
| `UNRAID_VALIDATE_QUERY` | `query { me { id name roles } }` | GraphQL query sent with the candidate key. Any 2xx with data validates; `UNAUTHENTICATED`/`FORBIDDEN` errors or HTTP 401/403 reject. Change only if your Unraid version lacks `me`. |
| `DATA_ROOT` | `/data` | Directory exposed by the file API. Its direct children are the shares. |
| `READ_ONLY` | `false` | Reject every mutating request with 403, regardless of mounts. |
| `LISTEN_ADDR` | `:8484` | Bind address. |
| `TLS_CERT` / `TLS_KEY` | | Serve HTTPS directly when both are set. Otherwise plain HTTP behind a proxy. |
| `SESSION_TTL` | `12h` | Lifetime of a session token issued at login. Clients re-login transparently on 401. |
| `MAX_LOGIN_ATTEMPTS` | `5` | Failed logins (or invalid `x-api-key` uses) per client IP before lockout. |
| `LOGIN_LOCKOUT` | `15m` | Lockout duration. A successful login resets the counter. |
| `USER_AUTH` | `optional` | `optional`: a login may carry an Unraid username and password on top of the API key; the session then has that user's share permissions. `required`: every login must. `off`: API key only. |
| `UNRAID_SMB_ADDR` | host of `UNRAID_URL` + `:445` | Unraid SMB endpoint used to verify user passwords (a real SMB2 session is opened and closed). |
| `SHARES_CONFIG` | `/unraid-shares/smb-shares.conf` | Unraid share security source, mounted read-only: the generated `/etc/samba/smb-shares.conf` (recommended, world-readable) or the `/boot/config/shares` directory (root-only on most systems). |
| `TRUST_PROXY` | `false` | Take the client IP from `X-Forwarded-For` / `X-Real-IP`. Set `true` behind Cloudflare, Nginx Proxy Manager, SWAG… so rate limiting sees the real client. Leave `false` when clients connect directly, or anyone could spoof the header. |
| `UPLOAD_TTL` | `24h` | How long an unfinished resumable upload (and its temp file) is kept. |
| `CHANGES_WALK_LIMIT` | `250000` | Maximum entries scanned by one `/fs/changes` page. |
| `CHANGES_DEADLINE` | `20s` | Maximum wall time of one `/fs/changes` page. On shfs expect roughly 3,000 entries/s. |
| `MAX_JSON_BODY` | `1048576` | Size cap for JSON request bodies and GraphQL queries (bytes). |
| `LOG_REQUESTS` | `true` | One JSON log line per request on stdout. |
| `TZ` | | Container timezone (log timestamps only; API timestamps are UTC). |

## Volumes

| Container path | Host path | Purpose |
|---|---|---|
| `/config` | `/mnt/user/appdata/unraid-gateway` | Item index (SQLite). Optional but recommended: without it the index lives in a temporary file and is rebuilt at every restart. |
| `/data/<share>` | `/mnt/user/<share>` | One mapping per share to expose (rw or ro). |
| `/unraid-shares/smb-shares.conf` | `/etc/samba/smb-shares.conf` | Share security for per-user access (read-only). |

### Legacy notes

Mount each share at `/data/<name>`. For per-user permissions also mount `/etc/samba/smb-shares.conf` → `/unraid-shares/smb-shares.conf` read-only (the Unraid template does this by default). `<name>` is what clients see; it may differ from the share name. Use `:ro` for read-only. Nothing outside `/data` is ever touched; the container needs no other mounts.

## User and permissions

The image runs as UID 99 / GID 100 (`nobody:users`), Unraid's defaults, so files created through the gateway carry the same ownership as files created over SMB and are visible to every Unraid share user. Files are created `0664`, directories `0775`. If your shares use different ownership, run the container with `--user <uid>:<gid>`.

## Web UI stored key (0.8+)

| Variable | Default | Meaning |
|---|---|---|
| `WEBUI_API_KEY` | empty | An Unraid API key the web UI may use with one click. |
| `WEBUI_API_KEY_FILE` | `/config/webui.key` | Where "Remember this key on the gateway" stores the key (0600). Empty disables remembering. |

Whoever can open the web UI can use the stored key: keep the gateway on the LAN or behind Cloudflare Access. The Unraid user/password (USER_AUTH) are still asked.
