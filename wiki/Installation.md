# Installation


## Community Applications and Homebrew

- **Community Applications**: the listing has been requested (see `docs/community-applications.md` in the repository). Once approved, *Apps › search "unraid-gateway"* installs the same template as the URL method below.
- **Homebrew** (macOS or Linux machine with the shares mounted, for example a Mac mini next to the NAS): `brew install sidimam/tap/unraid-gateway`, then `brew services start unraid-gateway`. Point `DATA_ROOT` at a folder with one sub-folder (or symlink) per share and `UNRAID_URL` at the Unraid WebGUI. The container remains the recommended way on Unraid itself.

## Unraid (recommended)

1. **Docker → Add Container → Template repositories**, add
   `https://raw.githubusercontent.com/sidimam/unraid-gateway/main/templates/unraid-gateway.xml`
   and save.
2. **Add Container → Template → unraid-gateway** (under *User templates*).
3. Set **Unraid WebGUI URL** to your server's LAN IP, e.g. `http://192.168.0.100`. Leave **User authentication** on `optional` and the **Unraid share security** mapping (`/etc/samba/smb-shares.conf`, read-only) in place so clients may log in with their Unraid user and get their SMB permissions.
4. Adjust the share **Paths**: each `/data/<name>` → `/mnt/user/<share>`, `Read/Write` or `Read Only`. Add or remove as needed. Never map `appdata`, `system`, `domains`, `isos` or Time Machine shares.
5. **Apply**. The *WebUI* button opens `http://<ip>:8484/`.

The template pins `ghcr.io/sidimam/unraid-gateway:latest`, so Unraid's *Check for Updates* works and *CA Auto Update Applications* can keep it current.

## docker-compose

```yaml
services:
  unraid-gateway:
    image: ghcr.io/sidimam/unraid-gateway:latest
    container_name: unraid-gateway
    restart: unless-stopped
    ports:
      - "8484:8484"
    environment:
      UNRAID_URL: "http://192.168.0.100"
      TRUST_PROXY: "true"
      USER_AUTH: "optional"
      TZ: "Europe/Rome"
    volumes:
      - /mnt/user/documents:/data/documents
      - /mnt/user/media:/data/media
      - /mnt/user/downloads:/data/downloads:ro
      - /etc/samba/smb-shares.conf:/unraid-shares/smb-shares.conf:ro
      - /tmp/notifications:/unraid-notifications      # Unraid notifications with any key role (0.11+)
```

## docker run

```bash
docker run -d --name unraid-gateway --restart unless-stopped \
  -p 8484:8484 \
  -e UNRAID_URL=http://192.168.0.100 -e TRUST_PROXY=true -e TZ=Europe/Rome -e USER_AUTH=optional \
  -v /mnt/user/documents:/data/documents \
  -v /mnt/user/media:/data/media \
  -v /etc/samba/smb-shares.conf:/unraid-shares/smb-shares.conf:ro \
  -v /tmp/notifications:/unraid-notifications \
  ghcr.io/sidimam/unraid-gateway:latest
```

## Verify

```bash
curl -s http://192.168.0.100:8484/healthz
# {"status":"ok","version":"v0.2.1"}
```

Then open the web UI, paste an Unraid API key (*Settings → Management Access → API Keys*, role `VIEWER` is enough) and browse.

## Image tags

| Tag | Meaning |
|---|---|
| `latest` | every push to `main`; what the Unraid template uses |
| `0.2.1`, `0.2` | releases (`v*` git tags) |
| `sha-<commit>` | exact commit |

Images are multi-arch (amd64, arm64) and pullable anonymously from GHCR.
