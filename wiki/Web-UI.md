# Web UI

Served at `/` by the gateway itself (embedded in the binary, no external assets).

- **Login**: paste an Unraid API key, optionally with an Unraid user and password. It is exchanged for a session token kept in the browser tab (`sessionStorage`); nothing is stored. With a user, the header shows the user name and read-only shares are marked at the root.
- **Header**: the key's name and role as reported by Unraid.
- **Shares**: the root lists the mounted shares. Open one to browse. Inside a share: **New folder**, **Upload** (with progress; one request per file, so mind proxy body limits), click a file to download, **Rename** / **Delete** on files and folders. Share roots have no rename/delete: they are mount points.
- **For the iOS app**: the URL to enter in Unraid Drive, with a note on whether the page is served over HTTPS (required by the Files app extension).
- **GraphQL proxy**: a text box to run queries through the proxy with the current key.

The UI is a convenience for testing and administration; the app talks to the JSON API directly.

## Activity panel (0.7+)

Below the file browser, once logged in. **Connected devices** lists every live session (or API key + IP): device and app build (from the `X-Unraid-Drive-Client` header the apps send; browsers show their User-Agent), Unraid user, IP, since / last seen, number of requests and the last file touched. A row with an orange edge has a transfer in flight. **Streams and transfers in progress** shows downloads, streams and uploads with the file, who is doing it, a progress bar with bytes and percent, the current speed and the elapsed time. **Recent transfers** keeps the last 50 finished ones. The panel polls `GET /api/v1/activity` every 3 seconds while the *live* box is ticked and the tab is visible. Scope: with an Unraid user you see only your own activity; an API-key-only login or a key with the ADMIN role sees everyone. Nothing is stored on disk.
