# Devices, notifications and API keys (0.9+)

## Devices

Every Unraid Drive installation (iPhone, iPad, Mac, Vision Pro, Apple TV — and a browser using the web UI) registers itself the first time it signs in: the app sends a stable installation id and its description, the gateway stores it in `/config/devices.json`. After that, every session belongs to a device.

**Web UI → Devices** lists them: device (app version, device family, OS, App / File Provider / Apple TV), Unraid user, API key name, registered, last seen, IP, number of logins. **Remove** closes that device's sessions on the spot; the app receives "this device was removed from the gateway", forgets its session and shows *Sign in again*: only that explicit action registers the device again (and raises a notification). **Remove all** does it for every device. From the container console: `gw devices`, `gw devices rm <id>`, `gw devices rm all`.

Who sees what: a session opened with an Unraid user sees and removes only its own devices; an API-key-only session or an ADMIN key manages all of them. `DEVICE_REGISTRATION=off` turns the registry off. Apps older than 1.3 build 29 send no device id: they keep working but are not listed.

### Reinstalled or replaced devices (0.9.1+)

A reinstalled app usually comes back with the **same** id (Unraid Drive 1.3 build 30 keeps it in iCloud per hardware) and simply refreshes its entry. When it cannot (all the developer's apps were removed, a new phone with the same name), it registers a new id: the gateway then marks the older entry with the same name and Unraid user as **old** — greyed in the web UI with *last seen …*, `OLD` in `gw devices`, `supersededBy`/`supersededAt` in the JSON. Remove it when you like; if that old installation ever signs in again, the flag is cleared.

## Notifications

A **new device** (warning) and a **removed device** (info) are announced on every configured channel:

- **Unraid** (`NOTIFY_UNRAID=true`, default): the gateway creates a notification through the Unraid API with the key of the session that registered the device. If Unraid refuses (a VIEWER key may lack the permission), set `NOTIFY_UNRAID_API_KEY` to an ADMIN key used only for this.
- **E-mail**: `SMTP_HOST`, `SMTP_PORT` (587), `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`, `SMTP_TO` (comma-separated), `SMTP_TLS` = `starttls` | `tls` | `none`. Gmail: `smtp.gmail.com`, 587, starttls, an app password.
- **Telegram**: create a bot with @BotFather, put its token in `TELEGRAM_BOT_TOKEN`, write to the bot once and read your chat id (for example from `https://api.telegram.org/bot<token>/getUpdates`), set `TELEGRAM_CHAT_ID`.

**Send a test** in the web UI (Notifications card) or `gw notify-test` reports each channel's result.

## Unraid API keys

With an **ADMIN** key the web UI card *Unraid API keys* talks to the Unraid API for you: **List keys**, **Create** (name + VIEWER/ADMIN), **Delete**, and **Rotate** — creates a replacement key, optionally remembers it for the web UI's one-click login and optionally deletes the old key right away. A new key is shown **once**: paste it into the apps (server → *Edit server or credentials*) and, if you rotate, do it before deleting the old key or every app stops until updated. With a VIEWER key Unraid refuses these operations and the card shows the error.

API: `GET /api/v1/devices`, `DELETE /api/v1/devices/{id}`, `DELETE /api/v1/devices`, `GET /api/v1/notify/channels`, `POST /api/v1/notify/test`, `GET /api/v1/keys`, `POST /api/v1/keys`, `DELETE /api/v1/keys/{id}`, `POST /api/v1/keys/rotate`. Login accepts `deviceId`, `deviceName`, `registerDevice`; answers carry `code: device_not_registered` (403 at login) or `device_revoked` (401 on a revoked session).
