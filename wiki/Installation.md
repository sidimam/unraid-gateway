# Installation

## Unraid (recommended)

1. **Docker → Add Container → Template repositories**, add
   `https://raw.githubusercontent.com/sidimam/unraid-gateway/main/templates/unraid-gateway.xml`
   and save.
2. **Add Container → Template → unraid-gateway** (under *User templates*).
3. Set **Unraid WebGUI URL** to your server's LAN IP, e.g. `http://192.168.0.100`.
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
      TZ: "Europe/Rome"
    volumes:
      - /mnt/user/documents:/data/documents
      - /mnt/user/media:/data/media
      - /mnt/user/downloads:/data/downloads:ro
```

## docker run

```bash
docker run -d --name unraid-gateway --restart unless-stopped \
  -p 8484:8484 \
  -e UNRAID_URL=http://192.168.0.100 -e TRUST_PROXY=true -e TZ=Europe/Rome \
  -v /mnt/user/documents:/data/documents \
  -v /mnt/user/media:/data/media \
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
