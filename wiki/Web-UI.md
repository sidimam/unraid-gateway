# Web UI

Served at `/` by the gateway itself (embedded in the binary, no external assets).

- **Login**: paste an Unraid API key, optionally with an Unraid user and password. It is exchanged for a session token kept in the browser tab (`sessionStorage`); nothing is stored. With a user, the header shows the user name and read-only shares are marked at the root.
- **Header**: the key's name and role as reported by Unraid.
- **Shares**: the root lists the mounted shares. Open one to browse. Inside a share: **New folder**, **Upload** (with progress; one request per file, so mind proxy body limits), click a file to download, **Rename** / **Delete** on files and folders. Share roots have no rename/delete: they are mount points.
- **For the iOS app**: the URL to enter in Unraid Drive, with a note on whether the page is served over HTTPS (required by the Files app extension).
- **GraphQL proxy**: a text box to run queries through the proxy with the current key.

The UI is a convenience for testing and administration; the app talks to the JSON API directly.
