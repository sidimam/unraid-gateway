# Reverse proxies and Cloudflare

The gateway speaks plain HTTP on 8484 and expects TLS to be terminated in front of it. iOS File Provider extensions refuse self-signed certificates, so the public URL must carry a certificate a stock device trusts.

## Cloudflare Tunnel (recommended)

- Public hostname → service `http://<unraid-ip>:8484`. No port forwarding, works behind CGNAT.
- Set `TRUST_PROXY=true`: Cloudflare sends `CF-Connecting-IP` / `X-Forwarded-For`, and the gateway's lockouts must apply to the real client, not to Cloudflare's IP.
- **Request body limit**: 100 MB per request on Free/Pro. The app uploads in 8 MB chunks; the web UI's one-shot upload is limited to 100 MB per file.
- Optional **Cloudflare Access** in front: create a Service Token and a *Service Auth* policy; the app sends `CF-Access-Client-Id` / `CF-Access-Client-Secret` on every request. The gateway needs no change.

## Nginx Proxy Manager / SWAG / Caddy / Traefik

Proxy `https://gw.example.com` → `http://<unraid-ip>:8484` with a Let's Encrypt certificate. Recommended proxy settings:

- forward `X-Forwarded-For`, `X-Real-IP` (default in NPM/SWAG) and set `TRUST_PROXY=true`;
- **no request body size limit**, or at least large: nginx `client_max_body_size 0;` (NPM: *Custom Nginx Configuration*), otherwise whole-file `PUT`s from the web UI fail with 413. The app's chunked uploads survive any limit above 8 MB;
- proxy read/send timeouts of several minutes for large downloads: `proxy_read_timeout 600s;`;
- HTTP/2 on, websockets not needed;
- do **not** add basic auth: the app cannot answer it (use Cloudflare Access instead if you want a second gate).

## Direct TLS

Set `TLS_CERT` and `TLS_KEY` to PEM files mounted into the container and the gateway serves HTTPS itself. You still need a publicly trusted certificate.

## Header cheat sheet

| Header | Direction | Purpose |
|---|---|---|
| `Authorization: Bearer` | client → gateway | session token |
| `x-api-key` | client → gateway | alternative to a session |
| `X-Forwarded-For` / `X-Real-IP` | proxy → gateway | real client IP when `TRUST_PROXY=true` |
| `CF-Access-Client-Id` / `-Secret` | client → Cloudflare | Access service token, consumed by Cloudflare |
| `Range`, `If-Match`, `If-None-Match`, `X-Mtime`, `Upload-Offset` | client → gateway | see [API reference](API-Reference) |
