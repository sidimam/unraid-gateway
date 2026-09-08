# API reference

Base path: `/api/v1`. All responses are JSON unless noted; errors are `{"error":"message"}` with a meaningful status. Paths are absolute, `/`-separated, relative to the data root: `/documents/Invoices/2026-01.pdf`. The first component is a share.

## Authentication

```http
POST /api/v1/auth/login
Content-Type: application/json
{"apiKey":"<unraid api key>"}
```
```json
{"token":"…","expiresAt":"2026-09-09T02:00:00Z","identity":{"name":"unraid gateway","roles":["VIEWER"]},"readOnly":false,"version":"v0.2.1"}
```

Then send `Authorization: Bearer <token>`. Alternatively send `x-api-key: <unraid api key>` on each request; validations are cached for five minutes. `401` means the token expired (log in again), `429` means the client IP is locked out.

| Method | Path | |
|---|---|---|
| `POST` | `/auth/logout` | invalidates the token |
| `GET` | `/auth/session` | identity and `readOnly` |
| `GET` | `/info` | version, `readOnly`, identity, feature list |
| `GET` | `/healthz` (no auth) | `{"status":"ok","version":"…"}` |

## Entries

```json
{"name":"IMG_0001.HEIC","path":"/media/Photos/IMG_0001.HEIC","type":"file","size":2345678,
 "mtime":"2026-09-08T14:03:11.120Z","etag":"\"23cace-1856f2c0a3b1c2d0\"","mode":"-rw-rw-r--"}
```

`type` is `file` or `dir`. `etag` is derived from size and mtime; it is also sent as the `ETag` header on `stat`, `content` and `PUT`.

## Files

| Method | Path | Notes |
|---|---|---|
| `GET` | `/fs/list?path=&hidden=1` | Directory listing, directories first, case-insensitive order. Dot-files hidden unless `hidden=1`. Response has the directory `etag`/`mtime` and `entries`. |
| `GET` | `/fs/stat?path=` | One entry. |
| `GET` / `HEAD` | `/fs/content?path=&download=1` | Streams the file. Supports `Range`, `If-None-Match`, `If-Range`, `If-Modified-Since`. `download=1` adds `Content-Disposition: attachment`. |
| `PUT` | `/fs/content?path=&overwrite=false` | Whole-file upload; body is the content. Written to a temp sibling then renamed atomically. `201` created / `200` replaced. Headers: `If-Match: <etag>` for optimistic locking (`412` on mismatch), `X-Mtime: <RFC 3339>` to preserve the client's modification time. `409` if the target is a directory or exists with `overwrite=false`. |
| `POST` | `/fs/mkdir` | `{"path":"/share/a/b","parents":true}` → `201` entry. |
| `POST` | `/fs/move` | `{"from":"/a/x","to":"/b/x","overwrite":false}` → `200` entry. Falls back to copy+delete across devices. |
| `POST` | `/fs/copy` | same body → `201` entry. Recursive. |
| `POST` | `/fs/delete` | `{"path":"/a/x","recursive":false}` → `204`. Non-empty directory without `recursive` → `409`. |
| `GET` | `/fs/changes?path=&since=&after=&cursor=&limit=` | Change feed, see [Change feed and sync](Change-Feed). |

### Status codes

| Code | Meaning |
|---|---|
| `400` | invalid path, destination inside source, bad parameters |
| `403` | read-only gateway or share, share root mutation (shares cannot be renamed/moved/deleted, nothing can be created at root), permission denied |
| `404` | not found; also *share not found* when the first path component is not a mounted share |
| `409` | exists / not empty / directory conflicts / upload offset mismatch |
| `412` | `If-Match` mismatch |
| `507` | no space left on the share |

## Resumable uploads

For large files and unreliable links. The temp file lives next to the destination, so commit is an atomic rename.

```http
POST   /fs/uploads                       {"path":"/media/clip.mov","size":734003200,"overwrite":false}
       → 201 {"id":"…","path":"…","offset":0,"expiresAt":"…"}      (header Upload-Offset: 0)
PATCH  /fs/uploads/{id}                  Upload-Offset: <current offset>   body: bytes
       → 204, Upload-Offset: <new offset>
HEAD   /fs/uploads/{id}                  → Upload-Offset: <current offset>   (resume after a drop)
POST   /fs/uploads/{id}/commit           X-Mtime: <RFC 3339, optional>       → 201 entry
DELETE /fs/uploads/{id}                  → 204 (abort)
```

A `PATCH` with the wrong `Upload-Offset` returns `409` and the correct offset in the header. Two concurrent `PATCH`es on the same upload return `409 upload busy`. If `size` was given, commit refuses (`400`) until all bytes arrived. Sessions expire after `UPLOAD_TTL`.

## GraphQL proxy

```http
POST /api/v1/graphql
Authorization: Bearer <token>
{"query":"{ array { state } shares { name free used } }"}
```

Forwarded verbatim to `UNRAID_URL/graphql` with the session's `x-api-key`; the response is returned unchanged. What you can query depends on the key's Unraid role. Introspection is disabled by Unraid; the fields below were validated on Unraid 7.3 / unraid-api 4.37 with a `VIEWER` key:

```graphql
{ array { state capacity { kilobytes { free used total } } disks { name status temp fsSize fsFree } parityCheckStatus { status progress running } }
  shares { name free used comment }
  docker { containers { id names state image autoStart isUpdateAvailable webUiUrl iconUrl } }
  info { os { hostname uptime release } cpu { brand cores threads } }
  notifications { overview { unread { total warning alert } } }
  metrics { cpu { percentTotal } memory { percentTotal used total } } }
```

## curl cheat sheet

```bash
GW=https://gw.example.com
TOKEN=$(curl -s -X POST $GW/api/v1/auth/login -H 'Content-Type: application/json' \
          -d '{"apiKey":"'"$KEY"'"}' | jq -r .token)
H="Authorization: Bearer $TOKEN"
curl -s "$GW/api/v1/fs/list?path=/documents" -H "$H" | jq .
curl -s -T ./file.pdf "$GW/api/v1/fs/content?path=/documents/file.pdf" -H "$H"
curl -s "$GW/api/v1/fs/content?path=/documents/file.pdf" -H "$H" -o file.pdf
curl -s -X POST $GW/api/v1/fs/mkdir -H "$H" -d '{"path":"/documents/New"}'
curl -s "$GW/api/v1/fs/changes?path=/documents&since=0" -H "$H" | jq '{cursor,truncated,next,dirs:(.dirs|length),files:(.files|length)}'
```
