# Architecture

## The problem

Unraid's own API (GraphQL, built in since 7.2) manages the server but has **no file operations**. Files are reachable through SMB, NFS or SSH/SFTP, none of which is sensible to expose to the Internet, and none of which speaks the HTTP+TLS that an iOS File Provider extension can use through Cloudflare or a reverse proxy. SMB in the Files app only works on the LAN or over a VPN.

## The solution

One small container that:

1. **Authenticates with the Unraid API key.** On login the gateway sends the key to `http://<unraid>/graphql` and checks that Unraid accepts it. It then issues its own random session token (32 random bytes, in memory, expires after `SESSION_TTL`). The API key is never stored by the gateway.
2. **Serves files** from `/data`, into which you mount the shares you want, with an HTTP API designed for sync clients: listing with ETags, streaming download with `Range`, atomic whole-file `PUT`, resumable chunked uploads, `mkdir` / `move` / `copy` / `delete`, and a paginated change feed.
3. **Proxies GraphQL** to Unraid using the caller's API key, so the app can render a dashboard without the Unraid WebGUI being exposed.
4. **Serves a web UI** on `/` for testing and administration.

```
client ──HTTPS──▶ Cloudflare / reverse proxy ──HTTP──▶ unraid-gateway :8484
                                                          │        │
                                                   /data/<share>   http://<unraid>/graphql
                                                   (bind mounts)   (key validation + proxy)
```

## Design decisions

- **Mount-based authorization.** *What* is accessible is decided by Docker volume mounts (`rw` / `ro`), not by the key's role. This keeps the model trivially auditable: `docker inspect` tells you exactly what any key can reach.
- **Share roots are immutable.** `/data/<share>` are mount points. The API refuses to rename, move, replace or delete them and refuses to create anything at root level, because `os.RemoveAll` on a mount point would wipe the share's contents.
- **No database.** Sessions and upload sessions live in memory. A restart logs everyone out; clients re-authenticate silently.
- **No inotify.** Unraid's user shares are a FUSE filesystem (`shfs`); inotify does not see changes made over SMB. The change feed walks the tree comparing mtimes instead, in bounded pages. See [Change feed and sync](Change-Feed).
- **Atomic writes.** Uploads go to a temporary sibling (`.name.gwpart*`) and are renamed into place, so readers never observe a partial file and a crashed upload leaves no half file under the real name.
- **Standard library only.** No third-party Go modules: small image, small attack surface, trivial audits.

## What it deliberately does not do

- It does not manage users or ACLs: everyone with a valid key gets the same view. Run one gateway per user if you need different views.
- It does not implement WebDAV or SMB.
- It does not cache or index file contents.
- It does not terminate TLS by default (it can, with `TLS_CERT`/`TLS_KEY`), because a reverse proxy or Cloudflare does that better.
