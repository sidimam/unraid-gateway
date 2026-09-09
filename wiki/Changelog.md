# Changelog

## v0.5.0 (2026-09-09)
- Persistent item index (SQLite, `INDEX_DB`, mount `/config`): stable ids on every entry (`id`, `parentId`), change journal served by `GET /fs/changes?seq=`, `GET /fs/item?id=`, background scanner (`INDEX_DIR_SCAN` 5m, `INDEX_FULL_SCAN` 6h), write-through from every API write, on-demand reconcile on listings. The legacy mtime feed stays for older clients.
- Template: new `/config` path and index settings; `<Changes>`/`<Date>` for Community Applications.

## 2026-09-09
- Homebrew formula `sidimam/tap/unraid-gateway` (builds v0.4.0 from source, `brew services` unit).
- Community Applications listing prepared: template `<Changes>`/`<Date>`, submission notes in `docs/community-applications.md`.
- Documented that Appdata Backup restarts the container nightly unless the container is skipped in the plugin.

## v0.4.0 — 2026-09-08
- Readable logs by default: startup banner with the effective configuration, one aligned line per event with method, endpoint, status, duration, file touched, user and IP; explicit login messages; health checks hidden. `LOG_FORMAT`, `LOG_LEVEL`, `LOG_HEALTHCHECKS`.
- Console walkthrough: the container's Console opens with a guided status check; `gw status|shares|users|login|key|env|help`.

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
