# Changelog

## v0.3.2 — 2026-09-08
- Shares are the real mount points under the data root (from `/proc/self/mountinfo`): directories left behind in the root volume after a mapping was removed no longer show up as shares, and paths into them return 404.

## v0.3.0 — 2026-09-08
- Per-user access: optional (or required) Unraid username + password on login, verified against Unraid's SMB service; per-share permissions derived from Unraid's generated Samba configuration (public/secure/private, read/write lists) and enforced on every endpoint; guest-mapped logins rejected via a share probe. Login/session responses report `user` and `shares`.
- New settings `USER_AUTH`, `UNRAID_SMB_ADDR`, `SHARES_CONFIG`; template mounts `/etc/samba/smb-shares.conf` read-only.
- Web UI: optional user/password fields, read-only badge on shares.
- First external dependency: `github.com/hirochachacha/go-smb2` (pure Go SMB2 client).

## v0.2.1 — 2026-09-08
- Share roots are immutable: rename/move/replace/delete of `/data/<share>` and creation at root level return 403.
- Writes into a share that does not exist return 404 instead of creating a top-level directory.
- Web UI hides rename/delete at root level.

## v0.2.0 — 2026-09-08
- Embedded web UI: login, browse, upload with progress, download, rename, delete, GraphQL playground.
- README walkthrough (API key creation, curl examples).

## v0.1.0 — 2026-09-08
- Paginated change feed (`after` / `next` / `cursor`), `CHANGES_DEADLINE`.
- `EROFS` → 403 *share is mounted read-only*, `ENOSPC` → 507.

## Initial release — 2026-09-08
- Unraid API-key login with session tokens and per-IP lockout.
- File API: list, stat, streaming download with Range, atomic PUT, resumable uploads, mkdir/move/copy/delete, change feed.
- GraphQL proxy. Dockerfile (alpine, `nobody:users`), Unraid template, GitHub Actions multi-arch build to GHCR.
