# Change feed and sync

## Item index and journal (gateway 0.5+)

Since 0.5.0 the gateway keeps a persistent **item index**: a SQLite file (`INDEX_DB`, default `/config/index.db`, mount `/mnt/user/appdata/unraid-gateway` there) with one row per file and directory, a stable random id for each, and a **journal** of changes numbered by a growing `seq`. This is the same idea as the file cache behind OneDrive, Google Drive or Nextcloud clients.

How it stays correct:

- every write through the API (upload, folder, rename, move, copy, delete) is journaled immediately;
- listing a directory reconciles that directory with the disk on the spot, so what a client sees is what the index knows;
- a background scanner catches changes made behind the gateway's back (SMB, other containers): every `INDEX_DIR_SCAN` (5 min) it stats directories only and re-lists those whose modification time changed (that is what adding, removing or renaming an entry does), and every `INDEX_FULL_SCAN` (6 h) it re-stats every entry to catch edits that leave the folder date untouched.

Clients call `GET /fs/changes?seq=<last seq>` and get exactly the items created, modified, moved or deleted since then, with their ids, in milliseconds regardless of the tree size. `seq=0` (first sync) or a seq older than the retained journal (30 days) returns `reset:true`: the client enumerates again and continues from the returned `seq`. Renames in place keep the id (same inode); moves through the API keep the id of the whole subtree; moves done over SMB across directories appear as delete + create.

The index is disposable: delete the file and it is rebuilt at the next start (a full scan of 400,000 entries on shfs takes a few minutes; the gateway serves requests meanwhile).

## Legacy feed (gateways before 0.5, or `INDEX_DB=off`)

### Why not inotify

Unraid user shares are served by `shfs`, a FUSE filesystem. inotify on `/mnt/user` does not report changes made through SMB or by other containers, so a watcher would miss most edits. The gateway therefore offers a **scan-based change feed** that clients call periodically.

## Semantics

`GET /fs/changes?path=<dir>&since=<unix ns>` walks `<dir>` recursively and returns:

- `dirs`: directories whose **mtime** is newer than `since`. A directory's mtime changes when an entry is **added, removed or renamed** inside it. Clients re-list these directories and diff against what they knew to detect deletions and renames.
- `files`: regular files whose mtime is newer than `since` (created or modified). Their full entry is included.
- `cursor`: the value to store as the next `since`. It is taken **before** the walk, minus two seconds, so writes racing the scan are reported again rather than lost.
- `scanned`: entries examined on this page.

`since=0` returns everything (a full scan). Dot-files and gateway temp files (`*.gwpart*`) are excluded.

## Pagination

Large shares cannot be walked in one request: on shfs expect roughly 3,000 entries per second. A page stops after `CHANGES_WALK_LIMIT` entries or `CHANGES_DEADLINE` and returns `truncated: true` with `next` = the last path scanned, in walk order. The client continues with

```
GET /fs/changes?path=<dir>&since=<same>&after=<next>&cursor=<cursor from page 1>
```

`after` resumes exactly where the walk stopped (component-wise path comparison, so `a` and `a-b` sort like `filepath.WalkDir`, and the resume directory's own children are still visited). Entries are never reported twice. Passing `cursor` keeps the first page's cursor for the whole scan. Store `cursor` as the new `since` only after a page with `truncated: false`.

Measured on a real share with 403,693 entries: 6 pages, 117 seconds, zero duplicates, 117 MB peak memory in the container.

## How Unraid Drive uses it

The File Provider extension encodes `cursor|after` in the sync anchor. For a directory enumerator it requests changes for that directory; if the directory itself is in `dirs` it re-lists it and reports vanished names as deletions. The working-set enumerator scans from `/` and re-lists up to 40 changed directories it has seen before. `truncated` maps to `moreComing`.
