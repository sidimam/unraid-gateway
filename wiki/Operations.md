# Operations

## Health

`GET /healthz` → `{"status":"ok","version":"…"}`. The image defines a Docker `HEALTHCHECK` on it; the Unraid template adds `--health-cmd` so the Docker tab shows *healthy*.

## Logs

*Docker → unraid-gateway → Logs*. The log opens with a banner of the effective configuration (listen address, Unraid API URL, mounted shares, user authentication mode, lockout settings), then one readable line per event:

```
2026-09-08 20:43:57 INFO  login ok (user)                  user=sdimambro key="unraid gateway" ip=203.0.113.7 shares="documents=rw media=ro"
2026-09-08 20:43:57 INFO  login ok (api key only)          key="unraid gateway" ip=203.0.113.7
2026-09-08 20:43:57 WARN  login failed: Unraid rejected the API key  ip=203.0.113.7
2026-09-08 20:43:57 WARN  login failed: Unraid rejected the username/password  user=bob ip=203.0.113.7 guest=false
2026-09-08 20:43:57 WARN  login refused: too many failed attempts from this IP  ip=203.0.113.7
2026-09-08 20:43:57 INFO  PUT    /api/v1/fs/content           → 201 in 42ms  file=/documents/report.pdf  user=sdimambro  ip=203.0.113.7
2026-09-08 20:43:57 INFO  GET    /api/v1/fs/list              → 404 in 0ms   file=/private  user=bob  ip=203.0.113.7
```

Access lines show the HTTP method and endpoint, the status and duration, the file or folder touched (`file=`), who did it (`user=` for an Unraid user, `user=key:<name>` for API-key-only sessions) and the client IP. Health-check probes are hidden unless `LOG_HEALTHCHECKS=true`. `LOG_FORMAT=json` restores one JSON object per line for log collectors; `LOG_LEVEL=debug` adds detail.

### Log growth

The gateway writes to stdout only; nothing is stored inside the container. Docker keeps the output in a `json-file` log on the Docker vDisk and Unraid caps it by default (*Settings → Docker → Docker log rotation*, `max-size 50m`, `max-file 1`), so the log is truncated at 50 MB and cannot fill the container or the array. Check the effective limit and current size with:

```sh
docker inspect unraid-gateway --format '{{.HostConfig.LogConfig.Type}} {{.HostConfig.LogConfig.Config}}'
du -h "$(docker inspect unraid-gateway --format '{{.LogPath}}')"
```

Versions before 0.4.0 logged every 30-second health probe as a JSON line (about 3 MB per month); 0.4.0 hides probes by default, so a normal week of use stays well under 1 MB. Recreating or updating the container starts a fresh log.

## Console walkthrough

`gw activity` (0.7.1+) shows who is connected and what is streaming right in the console; `gw activity -w` keeps refreshing. The same data is in the web UI's Activity panel after login.

The **Console** button of the container in the Docker tab (or `docker exec -it unraid-gateway sh`) opens a shell that immediately prints a guided status check. The `gw` helper is available:

| Command | What it shows |
|---|---|
| `gw status` | API health and version, Unraid API reachability, every mounted share with rw/ro, whether the share config is readable and Unraid's SMB is reachable (needed for user logins) |
| `gw shares` | each mounted share with its Unraid security (Public / Secure / Private) and read/write user lists |
| `gw users` | the Unraid users that appear in the share lists |
| `gw login <user>` | interactive test: asks the API key and the user's password, shows the login response with the resulting per-share permissions (token masked) |
| `gw key` | interactive test with the API key only |
| `gw env` | the effective configuration variables |
| `gw help` | this walkthrough |

Nothing is configured from the console: settings live in the container's variables and path mappings in Unraid.

## Nightly restarts (Appdata Backup)

The *Appdata Backup* plugin stops and restarts every container it does not skip, and logs `unraid-gateway does not have any volume to back up! Skipping. Please consider ignoring this container.` The gateway keeps no state, so set it to **Skip** in the plugin's per-container settings: otherwise phones syncing during the restart see the gateway unreachable and pause their Files location until the app wakes them again.

## Item index

The catalogue lives in `/config/index.db` (SQLite, WAL mode). Observed on an Unraid 7.3 server with shares on shfs: the first scan of 418,000 entries in 67,000 directories took 13 minutes at about 40 % of one CPU core and 140 MB of RAM, and the gateway served requests normally meanwhile; the resulting database is about 200 MB for that tree (0.5.1; 0.5.0 also wrote a journal row per file and reached 330 MB, delete the file once after upgrading). A second full scan on the same server took under 10 minutes. Afterwards the directory scan every 5 minutes only stats directories and takes seconds; the full scan every 6 hours repeats the initial walk in the background. The log reports `index: full scan done` and `index: directory scan` with counts and timings. Deleting `index.db` (container stopped) is always safe: it is rebuilt at the next start and clients simply re-enumerate.

## Updating

The template uses `latest`; Unraid's *Check for Updates* compares digests with GHCR and *Update* recreates the container from the template. *CA Auto Update Applications* can automate it. Sessions are lost on update; clients log in again transparently.

## Backups

Nothing to back up: the gateway holds no state besides in-memory sessions and in-flight upload temp files (`.name.gwpart*` next to the destination, removed on commit/abort/expiry).

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| Container exits immediately | `UNRAID_URL` missing or without scheme; a mapped host path does not exist. Read the first log lines. |
| `invalid api key` for a good key | `UNRAID_URL` wrong (must be the LAN IP, `http://` unless the WebGUI forces HTTPS), Unraid API stopped (`unraid-api status`), or key deleted. Log shows `unraid validation error`. |
| `401 invalid unraid username or password` | Samba rejected the user (wrong password, or unknown user mapped to guest and failing the share probe). Reset the password in the Unraid WebGUI → Users. |
| `401 this Unraid user has no access to any share mounted in the gateway` | Unraid's share security grants the user nothing on the mapped shares: add the user to the share's read/write list, or map another share. |
| `502 the Unraid share configuration is not readable by the gateway` | Mount `/etc/samba/smb-shares.conf` → `/unraid-shares/smb-shares.conf:ro` (template default) or set `USER_AUTH=off`. |
| `429 too many failed attempts` | lockout; wait `LOGIN_LOCKOUT` or restart. Behind a proxy without `TRUST_PROXY` everyone shares one IP. |
| `413` on uploads from the web UI | proxy body limit (Cloudflare 100 MB, nginx default 1 MB). Raise it or use the app. |
| Writes fail with `share is mounted read-only` | the mount has `:ro`. |
| `404 share not found` on write | the first path component is not a mounted share (typo, or share not mapped). |
| Change feed always truncated | very large share: normal; clients page with `after`. Raise `CHANGES_DEADLINE` if you prefer fewer, longer pages. |
| Dates off by hours | display only: API times are UTC; set `TZ` for log timestamps. |

### "permission denied" although the share is read-write

The container runs as Unraid's `nobody:users` (uid 99, gid 100), like Samba does for files it creates. It can only write where that account may write. Unraid's convention is that everything on the shares is `nobody:users` with mode `0777`/`0666` — this is what **Tools › New Permissions** enforces and what the gateway itself creates (0.5.5+). Folders created **outside** Unraid Drive — over SSH, with `rsync`/`scp`, or by another container running as a different user — usually get `0755`/`0644` and lock the gateway out.

Since 0.5.5 the error names the culprit, for example:

```
permission denied: the gateway runs as nobody (uid 99) and may not write in "/programs/Unraid-Drive/wiki" (owner sdimambro (uid 1000), gid 100, mode 0755). On Unraid run Tools › New Permissions on this share, or chmod -R ugo+rwX the folder.
```

Fix: in the Unraid web UI run **Tools › New Permissions** on that share (or `chmod -R ugo+rwX` the folder). When copying with rsync over SSH, add `--chown=nobody:users --chmod=ugo=rwX`. The start-up log also warns for every read-write share the gateway cannot write to.
