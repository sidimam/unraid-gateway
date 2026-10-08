# Web UI

Served at `/` by the gateway itself (embedded in the binary, no external assets). Redesigned in 0.10.0; reorganised in 0.12.0 around three tabs — **Status**, **Files**, **Settings** — with the same logic as the Unraid Drive apps: the useful content first, one place for every setting, one health dot shared with the apps.

## Header
- **Health dot** (0.12): green = all fine, yellow = warnings, red = problems, grey while unknown; the text next to it says which, and hovering lists the reasons. The same traffic light the Unraid Drive apps show next to the server name, computed by `GET /api/v1/health` (see [API reference](API-Reference#health-and-system-012)).
- After login: the user and the key's name and roles as reported by Unraid, **Forget stored key** (when a key is remembered on the gateway) and **Disconnect**. Language and theme moved to *Settings › Appearance* in 0.12.

## Login
- **Account name** + **Unraid API key**: a real form (`username` + `current-password` fields), so Safari / iCloud Keychain, Chrome and Firefox offer to save the key on the first login and fill it afterwards. The account name is only a label for your password manager (default `unraid-gateway`; change it if you keep several gateways). Chromium browsers are also offered the credential through the Credential Management API.
- **Sign in as an Unraid user…** opens a second, optional form: Unraid user + password (saved separately by the password manager). The gateway then applies that user's share permissions exactly as over SMB; the API key is still required.
- **Remember this key on the gateway** (0.8+): stores the key in `/config/webui.key` for the one-click *Connect with the key stored on this gateway* button. Anyone who can open the page could use it — LAN or Cloudflare Access only.
- The key is validated by Unraid and exchanged for a session token kept in the tab (`sessionStorage`).

## Status (default tab)
- **Health** card: the dot, the level and the reasons behind a yellow or red (array not started, a disk not OK or too hot, alert or warning notifications, CPU or memory above 90 %, the unraid-gateway container not running, Unraid API unreachable; gateway side: no share mounted, a read/write share the container cannot write to, `/config` not writable, user auth without the share configuration). *All checks* expands the full list. Refreshed every 30 s.
- Four counters: **Connected now**, **Streams / transfers**, **Registered devices** (superseded entries excluded), **Gateway** version and host.
- **System information**: host, Unraid release and kernel, up since, processor (cores, threads, clock), installed memory and modules, mainboard, Unraid API version, GPUs and the network interfaces with an address — from `GET /api/v1/system`.
- **Activity** (0.7+): connected devices (device, user, IP, since / last seen, requests, last file), streams and transfers in progress (progress bar, speed, elapsed), recent transfers. Refreshes every 3 s while *live* is on. A session opened with an Unraid user sees only its own rows; API-key-only and ADMIN sessions see everyone.

## Files
The root lists the mounted shares (read-only ones are marked when an Unraid user is signed in). Inside a share: **New folder**, **Upload** (with progress; one request per file, so mind proxy body limits), click a file to download, **Rename** / **Delete**. Share roots have no rename/delete: they are mount points.

## Settings (everything configurable, in one place)
- **Appearance**: **Language** — *System* (follows the browser) or one of English, Italiano, Español, Français, Deutsch, 简体中文, العربية, the same seven languages as Unraid Drive (dates follow the locale, Arabic switches to right-to-left); **Theme** — *System*, *Light* or *Dark*. Both remembered in the browser.
- **Devices** (0.9+) — every Unraid Drive installation; **Remove** closes its sessions and the app asks to sign in again; **Remove all**. Superseded entries (a reinstalled app) are greyed as *old*. See [Devices, notifications and API keys](Devices-Notifications-and-API-Keys).
- **Unraid API keys** (0.9+, ADMIN key) — list, create, delete, **Rotate**.
- **Notifications** (0.9+) — configured channels and **Send a test**.
- **For the apps** — the URL to enter in Unraid Drive (with a note on HTTPS, required by the Files app extension) and, folded away, the GraphQL proxy box.

Tables scroll horizontally inside their card on narrow windows; the last tab you used is reopened at the next login.

The UI is a convenience for testing and administration; the apps talk to the JSON API directly.
