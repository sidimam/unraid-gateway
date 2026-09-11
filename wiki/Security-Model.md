# Security model

## Surface

Only the gateway port is exposed (through Cloudflare or a proxy). The Unraid WebGUI, SSH and SMB stay on the LAN. Everything the gateway can do is bounded by the **volume mounts** and by a **valid Unraid API key**.

## Controls

| Control | Detail |
|---|---|
| Authentication | Every endpoint except `/healthz` and `/auth/login` requires a session token or `x-api-key`. Keys are validated by Unraid itself; the gateway stores neither keys nor passwords. |
| Brute force | `MAX_LOGIN_ATTEMPTS` failures per client IP → `LOGIN_LOCKOUT`. Enable `TRUST_PROXY` behind a proxy so the right IP is locked. |
| Sessions | 32 random bytes, in memory, `SESSION_TTL`. A restart invalidates all. |
| Path confinement | Client paths are cleaned (`..` clamped at root), joined under `DATA_ROOT`, and symlinks are resolved; anything escaping the root is rejected. |
| Share roots | Mount points cannot be renamed, moved, replaced or deleted; nothing can be created at root; writes into a non-existent share return 404 instead of creating a directory. |
| Privilege | Runs as `nobody:users` (99:100), no capabilities, no privileged mode, only `/data` mounted. |
| Atomic writes | Temp sibling + rename; no partial files under real names. |
| Web UI | Same-origin, strict Content-Security-Policy, session token in `sessionStorage` only; no cookies, no CORS headers. |
| Dependencies | None outside the Go standard library. Image based on Alpine with `ca-certificates`, `tzdata`, `curl` (for the health check). |

## Per-user access

When a login carries an Unraid username and password (`USER_AUTH` optional or required), the gateway opens an SMB2 session to Unraid with those credentials: Samba is the only authority on passwords and the gateway stores nothing but the outcome. Because Samba maps unknown users to guest, the gateway additionally opens a share that only the real user may open (a private share the user is listed on); a guest gets access denied and the login fails. It then reads the share definitions Unraid generates for Samba (`/etc/samba/smb-shares.conf`) and applies the SMB security model per share: **public** everyone rw; **secure** everyone ro, write list rw; **private** write list rw, read list ro, others hidden; shares with `shareExport` not containing `e` hidden. The rules are enforced on every request (list, stat, download, upload, mkdir, move, copy, delete, change feed, resumable uploads). Requests into hidden shares answer 404, exactly like SMB.

## What the key's role does and does not do

Roles (`VIEWER`, `ADMIN`, …) are enforced by Unraid on the **GraphQL proxy** only. A `VIEWER` key can read the dashboard but cannot start containers; an `ADMIN` key can. **File access is identical for every valid key**: whatever the mounts allow. Use `VIEWER` unless you need the proxy to change things, and mount only what you need.

## Media tickets (0.6+)

`POST /api/v1/fs/ticket` is only reachable with normal credentials and checks read access to the file. It returns an HMAC-SHA256 token bound to that one path and an expiry (default 8 h, max 24 h) signed with a secret generated at start-up, so **a restart invalidates every ticket**. `GET /media/<ticket>` needs no header: it is meant for players that cannot send one (libmpv on Apple TV, VLC, a browser). Threat model: a leaked ticket grants **read-only access to that single file until it expires**, nothing else — no key, token or password is ever in the URL. Keep TTLs short when the gateway is exposed to the Internet; behind Cloudflare Access the `/media/*` path has to be excluded from the Access policy for such players to work, which is why LAN use is the safer default.

## Stored API key (0.8+)

`WEBUI_API_KEY` / `/config/webui.key` let the web UI log in without pasting the key. The key is stored in clear (0600, inside the container's /config); whoever can open the web UI can use it, with the key's role and, without USER_AUTH, the full mounts. Only enable it on a gateway that is not reachable from the Internet without Cloudflare Access.

## Devices (0.9+)

Every app installation is registered; revoking it closes its sessions and forces a new sign-in. This is the tool to cut off a lost phone: remove the device, then rotate the API key if the key itself may have leaked. See [Devices, notifications and API keys](Devices-Notifications-and-API-Keys).

## Activity panel (0.7+)

The web UI shows who is connected and what is being streamed. A session opened with an Unraid username sees only its own devices and transfers; API-key-only sessions and ADMIN keys see every user. The data lives in memory only (no history beyond the last 50 transfers, idle clients dropped after 30 minutes). If family members should not see each other's activity, give them Unraid users and keep the API key with the ADMIN/VIEWER role for yourself.

## Recommendations

1. One API key per device, named after it; revoke on loss.
2. `Read Only` mounts for libraries and backups.
3. Never mount `appdata`, `system`, `domains`, `isos` or Time Machine shares.
4. Keep `latest` and let Unraid update the container.
5. Cloudflare Access (service token) if you want requests filtered before they reach your home.
6. Watch the logs: every request is one JSON line with client IP and status; `login failed` and `login locked` lines are worth alerting on.

## Reporting

Please report security issues privately via GitHub (*Security → Report a vulnerability*) rather than in a public issue.
