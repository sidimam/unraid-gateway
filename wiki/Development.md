# Development

## Layout

```
cmd/unraid-gateway/      main: config → server → http.Server with graceful shutdown
internal/config/         environment parsing
internal/auth/           Unraid key validation (GraphQL), session store, per-IP limiter, client IP
internal/fsapi/          file API: safe path resolution, handlers, resumable uploads, change feed
internal/proxy/          GraphQL passthrough
internal/webui/          embedded static UI
internal/server/         routing, auth middleware, logging, security headers
templates/               Unraid Community Applications template
assets/                  icon (SVG source + PNG)
```

## Build and test

```bash
go test ./...            # unit tests: path safety, handlers, resumable uploads, change feed pagination, auth/limiter
go vet ./... && gofmt -l .
go run ./cmd/unraid-gateway   # needs UNRAID_URL and DATA_ROOT
docker build -t unraid-gateway .
```

Without Go installed: `docker run --rm -v "$PWD":/src -w /src golang:1.23-alpine go test ./...`.

## Releasing

- Every push to `main` builds and pushes `ghcr.io/sidimam/unraid-gateway:latest` (amd64 + arm64) after tests pass.
- A `vX.Y.Z` tag additionally pushes `X.Y.Z` and `X.Y` tags. The binary's `/healthz` version comes from `git describe`.
- The Unraid template points at `latest`; nothing else to update.

## Compatibility promise

Everything under `/api/v1` stays backward compatible. Breaking changes will go to `/api/v2` with a major version bump, so a `latest` container never breaks an older app.

## Contributing

Issues and pull requests are welcome. Keep the standard-library-only rule, add a test for every handler change, run `gofmt`.
