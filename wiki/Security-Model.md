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

## Recommendations

1. One API key per device, named after it; revoke on loss.
2. `Read Only` mounts for libraries and backups.
3. Never mount `appdata`, `system`, `domains`, `isos` or Time Machine shares.
4. Keep `latest` and let Unraid update the container.
5. Cloudflare Access (service token) if you want requests filtered before they reach your home.
6. Watch the logs: every request is one JSON line with client IP and status; `login failed` and `login locked` lines are worth alerting on.

## Reporting

Please report security issues privately via GitHub (*Security → Report a vulnerability*) rather than in a public issue.
