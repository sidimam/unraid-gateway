# Changelog

## v0.8.0 (2026-09-11)
- **Stored API key for the web UI.** Two ways: the `WEBUI_API_KEY` container variable (masked in the template), or tick **Remember this key on the gateway** when you log in — the key goes to `WEBUI_API_KEY_FILE` (default `/config/webui.key`, mode 0600, so it survives container updates as long as /config is mapped). The login card then offers **Connect with the key stored on this gateway**; the Unraid user and password, when you use them, are still typed. **Forget stored key** in the header deletes the file. `GET /api/v1/auth/stored-key` tells the UI what is available; `POST /api/v1/auth/login {"useStoredKey": true}` uses it; `POST/DELETE /api/v1/auth/remember` store/forget it. Security: anyone who can open the web UI can use the stored key, so this is for a LAN-only or Cloudflare-Access-protected gateway.
- Console: **`gw activity`** prints who is connected (device, user, IP, since/last seen, requests, last file), what is streaming or uploading (with bytes, percent and elapsed time) and the recent transfers; `gw activity -w` refreshes every 3 s. Behind it, `GET /api/v1/activity?format=text` renders plain text, and requests from the container's own loopback address are served without a token (the raw socket address is checked, never proxy headers). The web UI login card now says what you get after connecting.

## v0.7.0 (2026-09-11)
- **Activity panel** in the web UI (after login): *Connected devices* — one row per session or API key + IP with device (the apps send `X-Unraid-Drive-Client: Unraid Drive 1.3 (28) · iPhone 17 Pro · iOS 26 · App/File Provider`; browsers and older builds show their User-Agent), Unraid user, IP, since / last seen, request count, last file touched; *Streams and transfers in progress* — downloads, streams (Range requests, media tickets) and uploads with file, who, progress bar, bytes, speed and elapsed time; *Recent transfers* (last 50). Refreshes every 3 s (toggle "live").
- `GET /api/v1/activity` returns the same data as JSON. Scope: a session opened with an Unraid user sees only that user's rows; API-key-only sessions and keys with the ADMIN role see everyone. Nothing is persisted: idle clients disappear 30 minutes after their last request, and a restart clears the lists.
- Transfers are counted at the byte level (response writer / request body wrappers), so progress is real, not estimated.

## v0.6.0 (2026-09-11)
- **Media tickets.** `POST /api/v1/fs/ticket {"path": "/media/movie.mkv", "ttl": "2h"}` (normal authentication, read access checked) answers `{"ticket", "url": "/media/<ticket>", "expiresAt"}`. `GET|HEAD /media/<ticket>` streams that one file with ETag and Range support and needs no header: it is how Unraid Drive on Apple TV feeds libmpv, and it works for VLC or a browser too. Tickets are HMAC-SHA256 signed with a per-process random secret (they expire on restart), bound to the file, default TTL 8 h, maximum 24 h. Nothing secret is in the URL; a leaked ticket gives read access to that single file until it expires. Behind Cloudflare Access the `/media/*` path must be excluded from the Access policy for header-less players (or use the gateway on the LAN).

## v0.5.5 (2026-09-11)
- Files and folders created through the gateway get Unraid's standard permissions: owner `nobody:users`, mode `0777` for folders and `0666` for files (what Tools › New Permissions sets). The process now runs with umask 0; before, the container umask reduced them to `0755`/`0644`, which locked other Unraid users (over SMB) and other containers out of files created from the app.
- A `403 permission denied` now explains itself: the folder the gateway could not write, its owner and mode, and the fix (`Tools › New Permissions` on the share, or `chmod -R ugo+rwX`). Typical cause: folders created over SSH, rsync or by another container as a different user with a 755 mode. The same message is logged at `WARN`.
- Start-up warns for every read-write share the gateway account cannot write to.

## v0.5.2 (2026-09-09)
- New icon (artwork by the author, light and dark versions in `assets/`); the web UI uses the dark one.

## v0.5.1 (2026-09-09)
- First catalogue scan no longer writes a journal row per file (the database stays small; a 420,000-entry tree took 330 MB before, a few tens of MB now). Delete `/config/index.db` once after upgrading from 0.5.0 to rebuild it lean.

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
