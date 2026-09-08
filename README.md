<p align="center">
  <img src="assets/icon.png" width="128" alt="unraid-gateway">
</p>

# unraid-gateway

**One authenticated port that lets a mobile app treat your Unraid shares like a cloud drive.**

`unraid-gateway` is a small Go service that runs as a Docker container on Unraid and exposes:

- a **file API** over the shares you choose: listing, streaming download with `Range`, atomic and *resumable* uploads, move/copy/delete, and a **change feed** designed for the iOS/iPadOS File Provider framework (the thing that makes a provider show up in the Files app next to iCloud Drive);
- a **proxy to the Unraid GraphQL API**, so a companion app can also read array status, shares, Docker containers and notifications without exposing the Unraid WebGUI itself.

Everything is authenticated with a regular **Unraid API key** (*Settings → Management Access → API Keys*). The gateway validates the key against Unraid on the LAN, hands the client a short-lived session token, and never stores your Unraid password.

It is the server half of a companion iOS app whose File Provider extension mounts your shares in the Files app. The gateway is protocol-agnostic and can be used by any HTTP client.

> No Unraid logo is used: the icon is an original design in the Unraid colour palette.

## Why not SMB / SFTP / WebDAV?

| | SMB | SFTP | WebDAV | unraid-gateway |
|---|---|---|---|---|
| Works from 5G / outside the LAN without VPN | ✗ | port-forward root SSH ✗ | ✓ | ✓ |
| Authenticates with the Unraid API key | ✗ | ✗ | ✗ | ✓ |
| Change feed for File Provider sync | ✗ | ✗ | ✗ | ✓ |
| Resumable uploads | ✗ | ✗ | ✗ | ✓ |
| Also proxies the Unraid GraphQL API | ✗ | ✗ | ✗ | ✓ |
| Needs a container on the server | ✗ | ✗ | ✓ | ✓ |

## Quick start (Unraid)

1. **Create an API key** in *Settings → Management Access → API Keys*. A dedicated key for the gateway is recommended.
2. **Install the container.** Until it is listed in Community Apps, add the template URL in *Docker → Add Container → Template repositories*:
   `https://raw.githubusercontent.com/sidimam/unraid-gateway/main/templates/unraid-gateway.xml`
   or run it with the [docker-compose.yml](docker-compose.yml) in this repo.
3. **Map only the shares you want on your phone.** Each path mapped under `/data/<Name>` appears as a top-level folder. Mark media shares read-only (`:ro`) if you like.
4. **Set `UNRAID_URL`** to the LAN address of your WebGUI, e.g. `http://192.168.1.10`. Use the IP, not `localhost` (the container has its own network namespace).
5. **Put TLS in front before exposing it.** iOS refuses self-signed certificates inside extensions. Use Nginx Proxy Manager or SWAG with Let's Encrypt, or Cloudflare Tunnel (no port forward needed, works behind CGNAT). Set `TRUST_PROXY=true` so rate limiting sees real client IPs.

```bash
curl -s https://gw.example.com/healthz
# {"status":"ok","version":"v0.1.0"}
```

## Configuration

| Variable | Default | Description |
|---|---|---|
| `UNRAID_URL` | **required** | LAN URL of the Unraid WebGUI (`http://192.168.1.10`). |
| `UNRAID_INSECURE_TLS` | `false` | Skip certificate verification when `UNRAID_URL` is https with a self-signed cert. |
| `UNRAID_VALIDATE_QUERY` | `query { me { id name roles } }` | GraphQL query used to validate API keys. |
| `DATA_ROOT` | `/data` | Directory exposed by the file API. |
| `READ_ONLY` | `false` | Disable every write operation. |
| `LISTEN_ADDR` | `:8484` | Bind address. |
| `SESSION_TTL` | `12h` | Lifetime of a login token. |
| `MAX_LOGIN_ATTEMPTS` | `5` | Failed logins per IP before lockout. |
| `LOGIN_LOCKOUT` | `15m` | Lockout duration. |
| `TRUST_PROXY` | `false` | Honour `X-Forwarded-For` / `X-Real-IP`. |
| `UPLOAD_TTL` | `24h` | How long an unfinished resumable upload is kept. |
| `CHANGES_WALK_LIMIT` | `250000` | Max entries scanned per `/fs/changes` call. |
| `TLS_CERT` / `TLS_KEY` | | Serve HTTPS directly instead of behind a proxy. |
| `LOG_REQUESTS` | `true` | Access log (JSON) on stdout. |

The container runs as `99:100` (`nobody:users`), so files created through the gateway get the same ownership as files created over SMB.

## API

All endpoints live under `/api/v1`. Responses are JSON unless noted. Errors are `{"error":"..."}` with a meaningful HTTP status.

### Authentication

```http
POST /api/v1/auth/login
{"apiKey":"<unraid api key>"}

200 {"token":"…","expiresAt":"…","identity":{"name":"root","roles":["admin"]},"readOnly":false}
```

Then send `Authorization: Bearer <token>` on every call. For scripts you may instead send `x-api-key: <unraid api key>` directly; the gateway caches validations for five minutes.

`POST /auth/logout`, `GET /auth/session`, `GET /info` are also available.

### Files

| Method | Path | Notes |
|---|---|---|
| `GET` | `/fs/list?path=/Photos&hidden=1` | Entries sorted dirs-first. Returns the directory `etag` too. |
| `GET` | `/fs/stat?path=` | Single entry. |
| `GET`/`HEAD` | `/fs/content?path=&download=1` | Streams the file. Supports `Range`, `If-None-Match`, `If-Range`. |
| `PUT` | `/fs/content?path=&overwrite=false` | Whole-file upload, written to a temp sibling and renamed atomically. `If-Match: <etag>` for optimistic locking, `X-Mtime: RFC3339` to preserve the client mtime. |
| `POST` | `/fs/mkdir` `{"path":"/a/b","parents":true}` | |
| `POST` | `/fs/move` `{"from":"/a","to":"/b","overwrite":false}` | Falls back to copy+delete across devices. |
| `POST` | `/fs/copy` `{"from":"/a","to":"/b"}` | Recursive. |
| `POST` | `/fs/delete` `{"path":"/a","recursive":false}` | Non-empty directory without `recursive` → `409`. |
| `GET` | `/fs/changes?path=/&since=<unix ns>` | Change feed, see below. |

Entry shape:

```json
{"name":"IMG_0001.HEIC","path":"/Photos/IMG_0001.HEIC","type":"file","size":2345678,
 "mtime":"2026-09-08T14:03:11.120Z","etag":"\"23cace-1856f2c0a3b1c2d0\"","mode":"-rw-rw-r--"}
```

### Resumable uploads

Designed for large files on flaky mobile links. The temp file lives next to the destination, so commit is an atomic rename on the same filesystem.

```http
POST   /fs/uploads                {"path":"/Video/clip.mov","size":734003200,"overwrite":false}
       → 201 {"id":"…","offset":0,"expiresAt":"…"}
PATCH  /fs/uploads/{id}           Upload-Offset: 0        <bytes…>   → 204, Upload-Offset: 8388608
HEAD   /fs/uploads/{id}           → Upload-Offset: <current>          (resume after a network drop)
POST   /fs/uploads/{id}/commit    X-Mtime: 2026-09-08T14:03:11Z       → 201 <entry>
DELETE /fs/uploads/{id}           → 204                               (abort)
```

A `PATCH` whose `Upload-Offset` does not match the server returns `409` with the correct offset. Unfinished uploads expire after `UPLOAD_TTL`.

### Change feed

`GET /fs/changes?path=/&since=<cursor>` walks the subtree and returns:

```json
{"cursor":1757340191000000000,
 "dirs":["/Photos/2026","/Documents"],
 "files":[{...},{...}],
 "truncated":false,"scanned":18234}
```

- `dirs` are directories whose mtime moved past `since`: something was **added, removed or renamed** inside them. Re-enumerate those.
- `files` are regular files modified after `since`. Refresh their metadata/content.
- Store `cursor` and pass it as `since` next time. `since=0` is a full scan. The cursor is deliberately two seconds in the past so writes racing the scan are reported again rather than lost.
- Files created by the gateway itself (`*.gwpart`) and dot-files are excluded.

This maps directly onto `NSFileProviderReplicatedExtension`'s enumerator + sync-anchor model without needing inotify on the shfs FUSE mount.

### GraphQL proxy

```http
POST /api/v1/graphql
Authorization: Bearer <token>
{"query":"{ array { state } shares { name free } }"}
```

Forwarded to `UNRAID_URL/graphql` with the session's `x-api-key`. Response is returned verbatim.

## Security model

- The Unraid API key is sent **once** at login over TLS; afterwards only an opaque random session token travels. Sessions live in memory and die with the container.
- Login is rate-limited per client IP with a temporary lockout. Enable `TRUST_PROXY` behind a reverse proxy or the proxy's IP will be what gets locked.
- Every path is confined to `DATA_ROOT`. `..` is clamped, symlinks pointing outside the root are refused.
- The container runs unprivileged as `nobody:users`; mount only the shares you need and use `:ro` where writes are not required.
- What the gateway exposes is decided by **volume mounts**, not by the API key's role. A `guest` key that Unraid accepts gets the same file access as an `admin` key. Create a dedicated key and treat it like a password.
- No CORS headers are sent: this is an API for native clients, not for browsers.

## Development

```bash
go test ./...
go run ./cmd/unraid-gateway   # needs UNRAID_URL, DATA_ROOT
docker build -t unraid-gateway .
```

Images for `linux/amd64` and `linux/arm64` are published to `ghcr.io/sidimam/unraid-gateway` by GitHub Actions on every push to `main` and on `v*` tags.

## Roadmap

- [ ] Companion iOS/iPadOS app with File Provider extension
- [ ] Optional per-share permissions bound to Unraid roles
- [ ] Thumbnails endpoint for images/videos
- [ ] Community Apps listing

## License

MIT
