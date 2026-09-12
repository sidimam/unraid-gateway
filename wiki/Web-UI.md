# Web UI

Served at `/` by the gateway itself (embedded in the binary, no external assets). Redesigned in 0.10.0.

## Header
- **Language**: *System* (follows the browser) or one of English, Italiano, Español, Français, Deutsch, 简体中文, العربية — the same seven languages as Unraid Drive. Dates follow the chosen locale; Arabic switches the layout to right-to-left. Remembered in the browser.
- **Theme**: *System* (follows `prefers-color-scheme`), *Light* or *Dark*. Remembered in the browser.
- After login: the user and the key's name and roles as reported by Unraid, **Forget stored key** (when a key is remembered on the gateway) and **Disconnect**.

## Login
- **Account name** + **Unraid API key**: a real form (`username` + `current-password` fields), so Safari / iCloud Keychain, Chrome and Firefox offer to save the key on the first login and fill it afterwards. The account name is only a label for your password manager (default `unraid-gateway`; change it if you keep several gateways). Chromium browsers are also offered the credential through the Credential Management API.
- **Sign in as an Unraid user…** opens a second, optional form: Unraid user + password (saved separately by the password manager). The gateway then applies that user's share permissions exactly as over SMB; the API key is still required.
- **Remember this key on the gateway** (0.8+): stores the key in `/config/webui.key` for the one-click *Connect with the key stored on this gateway* button. Anyone who can open the page could use it — LAN or Cloudflare Access only.
- The key is validated by Unraid and exchanged for a session token kept in the tab (`sessionStorage`).

## Dashboard
Four counters at the top: **Connected now**, **Streams / transfers**, **Registered devices** (superseded entries excluded), **Gateway** version and host. Then tabs:

- **Files** — the root lists the mounted shares (read-only ones are marked when an Unraid user is signed in). Inside a share: **New folder**, **Upload** (with progress; one request per file, so mind proxy body limits), click a file to download, **Rename** / **Delete**. Share roots have no rename/delete: they are mount points.
- **Activity** (0.7+) — connected devices (device, user, IP, since / last seen, requests, last file), streams and transfers in progress (with progress bar, speed, elapsed), recent transfers. Refreshes every 3 s while *live* is on. A session opened with an Unraid user sees only its own rows; API-key-only and ADMIN sessions see everyone.
- **Devices** (0.9+) — every Unraid Drive installation; **Remove** closes its sessions and the app asks to sign in again; **Remove all**. Superseded entries (a reinstalled app) are greyed as *old*. See [Devices, notifications and API keys](Devices-Notifications-and-API-Keys).
- **Notifications** (0.9+) — configured channels and **Send a test**.
- **API keys** (0.9+, ADMIN key) — list, create, delete, **Rotate**.
- **Advanced** — the URL to enter in the apps (with a note on HTTPS, required by the Files app extension) and the GraphQL proxy box.

Tables scroll horizontally inside their card on narrow windows; the last tab you used is reopened at the next login.

The UI is a convenience for testing and administration; the apps talk to the JSON API directly.
