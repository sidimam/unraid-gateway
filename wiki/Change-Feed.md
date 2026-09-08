# Change feed and sync

## Why not inotify

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
