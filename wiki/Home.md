# unraid-gateway — Wiki

`unraid-gateway` is a single container that runs on Unraid and exposes **one authenticated HTTPS port** for mobile clients: a streaming file API over the shares you choose, a proxy to the Unraid GraphQL API, and a small web UI. Authentication uses a regular **Unraid API key**.

It is the server side of **[Unraid Drive](https://github.com/sidimam/unraid-drive)**, the iPhone / iPad / Apple Vision Pro app that puts your shares in the Files app. End-user setup (container, API key, Cloudflare Tunnel, app) is documented step by step in the **[Unraid Drive wiki](https://github.com/sidimam/unraid-drive/wiki)**. This wiki is the technical reference for the gateway itself.

## Pages

- [Architecture](Architecture) — what the gateway does and does not do, and why
- [Installation](Installation) — Unraid template, docker-compose, environment variables
- [Configuration reference](Configuration) — every variable, default and effect
- [API reference](API-Reference) — authentication, file endpoints, resumable uploads, change feed, GraphQL proxy
- [Change feed and sync](Change-Feed) — how clients keep in sync without inotify
- [Reverse proxies and Cloudflare](Reverse-Proxies) — TLS, `TRUST_PROXY`, body limits, Cloudflare Access
- [Security model](Security-Model)
- [Web UI](Web-UI)
- [Operations](Operations) — logs, updates, health checks, troubleshooting
- [Development](Development) — building, testing, releasing
- [Changelog](Changelog)

## Quick facts

| | |
|---|---|
| Image | `ghcr.io/sidimam/unraid-gateway:latest` (linux/amd64, linux/arm64), about 10 MB |
| Language | Go 1.23, standard library plus a pure Go SMB2 client |
| Runs as | `nobody:users` (99:100) |
| Users | optional Unraid user + password on login, verified over SMB; per-share permissions from Unraid's Samba config |
| Port | 8484 (HTTP; put TLS in front) |
| Requires | Unraid 7.2+ with the built-in Unraid API enabled |
| License | MIT |
