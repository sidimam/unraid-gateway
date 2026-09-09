# Publishing unraid-gateway on Unraid Community Applications

Community Applications (CA) is the Unraid "app store". Templates are pulled from a public GitHub repository that the CA maintainers add to their feed after a one-time request. This repository already contains everything CA needs; the request itself must be made by the repository owner.

## What CA reads from this repository

- `templates/unraid-gateway.xml`: the container template (name, image, ports, paths, variables, overview, categories, `<Support>`, `<Project>`, `<Icon>`, `<TemplateURL>`, `<Changes>`).
- `assets/icon.png`: the application icon referenced by the template (not the Unraid logo, no animation).
- The source code (CA policy: templates made by their maintainer must point to open-source applications).

## One-time request (repository owner)

1. Enable **two-factor authentication** on the GitHub account that owns the repository (CA policy) and on the registry account (GHCR uses the GitHub account).
2. Fill in the CA inclusion form linked from the "CA Application Policies & Notes" thread on the Unraid forums (Squid's Asana form). Data to enter:
   - Template repository: `https://github.com/sidimam/unraid-gateway` (templates in the `templates/` folder).
   - Application name: `unraid-gateway`.
   - Docker image: `ghcr.io/sidimam/unraid-gateway:latest` (GitHub Container Registry, multi-arch amd64/arm64, public).
   - Support: `https://github.com/sidimam/unraid-gateway/issues` (or a forum support thread, see below).
   - Project: `https://github.com/sidimam/unraid-gateway`.
   - Short description: *Single authenticated port that serves your Unraid shares to the iOS/iPadOS/visionOS Files app (Unraid Drive) and proxies the Unraid API. Unraid API-key login, optional per-user share permissions, resumable uploads, change feed, web UI.*
3. Optional but recommended: open a support thread in the forum section *Docker Containers* titled `[Support] unraid-gateway` and put its URL in `<Support>` of the template; CA shows it as the "Support" link.
4. Allow up to 48 hours. After inclusion, users find **unraid-gateway** in Apps; template updates are picked up automatically from this repository (CA refreshes its feed regularly).

## Keeping the listing current

- Update `<Changes>` and `<Date>` in the template with every release: CA shows them in the app's "Changes" tab.
- Keep the image on the `latest` tag so Unraid's *Check for Updates* keeps working.
- Never change `<Name>` (it identifies the app in CA).

## Forum post draft

> **[Support] unraid-gateway** — the server side of the Unraid Drive app (iPhone, iPad, Apple Vision Pro).
>
> unraid-gateway exposes one authenticated port that the Files app on iOS can mount like iCloud Drive: list, download, resumable upload, move, rename, delete and a change feed, plus a proxy to the Unraid GraphQL API for the app's dashboard. Login uses your Unraid API key, optionally together with an Unraid username and password so each family member sees exactly the shares Unraid grants them. Publish it with Cloudflare Tunnel (works behind CGNAT) or any reverse proxy with TLS, optionally protected with Cloudflare Access.
>
> - Source and template: https://github.com/sidimam/unraid-gateway (MIT)
> - Setup guide: https://github.com/sidimam/unraid-gateway/wiki
> - The app: https://github.com/sidimam/unraid-drive (App Store)
>
> Please report problems here or on GitHub Issues.
