# Devices, notifications and API keys (0.9+)

## Devices

Every Unraid Drive installation (iPhone, iPad, Mac, Vision Pro, Apple TV — and a browser using the web UI) registers itself the first time it signs in: the app sends a stable installation id and its description, the gateway stores it in `/config/devices.json`. After that, every session belongs to a device.

**Web UI → Devices** lists them: device (app version, device family, OS, App / File Provider / Apple TV), Unraid user, API key name, registered, last seen, IP, number of logins. **Remove** closes that device's sessions on the spot; the app receives "this device was removed from the gateway", forgets its session and shows *Sign in again*: only that explicit action registers the device again (and raises a notification). **Remove all** does it for every device. From the container console: `gw devices`, `gw devices rm <id>`, `gw devices rm all`.

Who sees what: a session opened with an Unraid user sees and removes only its own devices; an API-key-only session or an ADMIN key manages all of them. `DEVICE_REGISTRATION=off` turns the registry off. Apps older than 1.3 build 29 send no device id: they keep working but are not listed.

### Device cards, metrics and log (0.13+)

**Web UI → Settings → Devices** shows one card per installation: an icon by kind (iPhone, iPad, Mac, Apple TV, Vision Pro, browser), the device name, its OS and app build, the Unraid user, the last sync and a state dot — green when active today, grey when older (or *old*, superseded), yellow when not seen for more than 30 days. Open a card for the **detail page**:

- **Facts**: device, user, key, registered, last sync, IP, logins, state, id.
- **Data moved**: totals for today, the last 7 and the last 30 days (bytes downloaded or streamed ↓ and uploaded ↑, with requests, downloads/streams and uploads), and a bar chart — per hour over the last 24 hours, per day over 7, 30 or 90 days.
- **Log**: the device's events, newest first — sign-ins (and refused ones: wrong key, device not registered), registration, revocation (its session closed because the device was removed), devices it removed, uploads that stopped before the end.
- **Remove** — the same as before (sessions closed, the app asks to sign in again).

The gateway records these in `/config/metrics.json`: per device, hourly buckets (requests, bytes up/down, streams, uploads, downloads) kept for 90 days and the last 200 events; the file is flushed every minute and at shutdown. The apps do nothing special: the authentication middleware counts the request and the transfer tracker adds the bytes when a stream, download or upload ends. Removing a device drops its metrics.

**Stale devices**: *Remove the devices not seen for 30 / 90 / 180 / 365 days* (below the cards), `gw devices prune <days>` in the console, or `POST /api/v1/devices/prune {"days": N}` — API-key-only or ADMIN sessions; the caller's own device is never removed; a notification lists what was removed. If one of those devices comes back, its app asks to sign in again and registers anew.

API: `GET /api/v1/devices/{id}/metrics?range=24h|7d|30d|90d` → `{series: [{t, requests, up, down, streams, uploads, downloads}], step, today, week, month, lastActivity}`; `GET /api/v1/devices/{id}/log?limit=N` → `{events: [{time, kind, detail, level}]}` (kinds: login, register, revoked, transfer, admin; levels info/warning/error). A user session sees only its own devices.

### Users and syncs (0.14)

Unraid Drive 1.5 can keep folders in sync and back Photos up through the gateway. Every pair the app defines is registered here, so the administrator sees — per user — what each device syncs and can stop it.

**Web UI → Settings → Users** lists the Unraid users (or API keys) that have registered devices, with the number of devices and syncs, when last seen, what moved today and a dot that turns yellow when a device has not been seen for 30 days or a sync is paused / had problems. Opening a user shows:

- **Devices**: the same cards as the Devices section, with **Remove** (closes sessions, drops the device's syncs and metrics; the app asks to sign in again).
- **Folder syncs and backups**: one row per pair — kind (folder sync, Photos backup, Downloads), direction (device → Unraid, Unraid → device, two-way), the device, the last run (when, files, bytes) and its state — with a **switch**: off pauses the pair *on the gateway*: the app's next report is refused (`409 sync_disabled`, with who paused it) and the app shows the pair as paused until it is switched on again here or from the console. The app cannot re-enable a pair paused on the gateway. **Files (N)** opens the list of the last 100 files the pair moved: time, path, direction (uploaded, downloaded, deleted, conflict), bytes, outcome (ok, conflict, failed, skipped) and detail.
- **Data moved**: totals for today, 7 and 30 days and a chart over 24 h / 7 / 30 / 90 days, all the user's devices summed.
- **Log**: the events of every device of the user, merged, newest first; sync events (added, paused, resumed, removed, each run with files / bytes / conflicts / errors) are among them.

Console: `gw profile` (every user), `gw profile <user>`, `gw syncs`, `gw syncs pause|resume|rm <id>`.

Storage: `/config/syncs.json` (pairs with counters, last run and recent files). The registry is owned by the apps: a pair deleted in the app disappears here too; a device removed here takes its pairs with it.

API (0.14):

| Method | Path | Notes |
|---|---|---|
| `GET` | `/api/v1/syncs` | The caller device's pairs; `?all=1`, `?device=ID`, `?user=NAME` for the web UI / console (what the caller may see). `?format=text`. |
| `PUT` | `/api/v1/syncs/{id}` | `{kind, localLabel, remotePath, mode, enabled}` — the owning device creates or updates its pair (the id is the app's UUID). `enabled` is kept when the pair was paused on the gateway. |
| `PATCH` | `/api/v1/syncs/{id}` | `{"enabled": bool}` — pause / resume (web UI, console, or the owning app). |
| `DELETE` | `/api/v1/syncs/{id}` | Removes the pair. |
| `POST` | `/api/v1/syncs/{id}/runs` | `{run: {at, durationMs, files, bytes, conflicts, errors, status: ok|partial|failed|skipped, error, trigger}, files: [{time, path, direction, bytes, outcome, detail}]}` → the stored pair; `409 {"code": "sync_disabled", "disabledBy"}` when paused. |
| `GET` | `/api/v1/users` | `{users: [{user, devices, syncs, pausedSyncs, lastSeen, today, warnings}]}`; `?format=text`. |
| `GET` | `/api/v1/users/{name}` | `{user, devices: [...], syncs: [...], today, week, month}`; `?format=text`. |
| `GET` | `/api/v1/users/{name}/metrics?range=24h\|7d\|30d\|90d` | The user's devices summed: `{series, step, today, week, month}`. |
| `GET` | `/api/v1/users/{name}/log?limit=N` | `{events: [{time, kind, detail, level, deviceId, deviceName}]}` merged, newest first. |

### Reinstalled or replaced devices (0.9.1+)

A reinstalled app usually comes back with the **same** id (Unraid Drive 1.3 build 30 keeps it in iCloud per hardware) and simply refreshes its entry. When it cannot (all the developer's apps were removed, a new phone with the same name), it registers a new id: the gateway then marks the older entry with the same name and Unraid user as **old** — greyed in the web UI with *last seen …*, `OLD` in `gw devices`, `supersededBy`/`supersededAt` in the JSON. Remove it when you like; if that old installation ever signs in again, the flag is cleared.

## Notifications

A **new device** (warning) and a **removed device** (info) are announced on every configured channel. **Unraid's own notifications need nothing (0.11+):** the template maps the host folder `/tmp/notifications` to `/unraid-notifications` and the gateway writes the notification file itself, with any API key role — if your container predates 0.11, add the Path once (Edit › Add another Path: container `/unraid-notifications`, host `/tmp/notifications`, read/write). Without the mapping the gateway falls back to the API mutation, which Unraid allows only to ADMIN keys (`NOTIFY_UNRAID_API_KEY`). Spool notifications do not trigger Unraid's external agents (e-mail, Pushover…): configure SMTP or Telegram below for those.

Channels:

- **Unraid** (`NOTIFY_UNRAID=true`, default): the gateway creates a notification through the Unraid API with the key of the session that registered the device. If Unraid refuses (a VIEWER key may lack the permission), set `NOTIFY_UNRAID_API_KEY` to an ADMIN key used only for this.
- **E-mail**: `SMTP_HOST`, `SMTP_PORT` (587), `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`, `SMTP_TO` (comma-separated), `SMTP_TLS` = `starttls` | `tls` | `none`. Gmail: `smtp.gmail.com`, 587, starttls, an app password.
- **Telegram**: create a bot with @BotFather, put its token in `TELEGRAM_BOT_TOKEN`, write to the bot once and read your chat id (for example from `https://api.telegram.org/bot<token>/getUpdates`), set `TELEGRAM_CHAT_ID`.

**Send a test** in the web UI (Notifications card) or `gw notify-test` reports each channel's result.

## Unraid API keys

With an **ADMIN** key the web UI card *Unraid API keys* talks to the Unraid API for you: **List keys**, **Create** (name + VIEWER/ADMIN), **Delete**, and **Rotate** — creates a replacement key, optionally remembers it for the web UI's one-click login and optionally deletes the old key right away. A new key is shown **once**: paste it into the apps (server → *Edit server or credentials*) and, if you rotate, do it before deleting the old key or every app stops until updated. With a VIEWER key Unraid refuses these operations and the card shows the error.

API: `GET /api/v1/devices`, `DELETE /api/v1/devices/{id}`, `DELETE /api/v1/devices`, `GET /api/v1/devices/{id}/metrics`, `GET /api/v1/devices/{id}/log`, `POST /api/v1/devices/prune` (0.13), `GET /api/v1/notify/channels`, `POST /api/v1/notify/test`, `GET /api/v1/keys`, `POST /api/v1/keys`, `DELETE /api/v1/keys/{id}`, `POST /api/v1/keys/rotate`. Login accepts `deviceId`, `deviceName`, `registerDevice`; answers carry `code: device_not_registered` (403 at login) or `device_revoked` (401 on a revoked session).
