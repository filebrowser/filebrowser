# File Browser Collabora/WOPI integration

This fork adds built-in Collabora Online support by making File Browser act as a WOPI host.

## Runtime configuration

Admin users can configure Collabora from the File Browser GUI: **Settings → Global Settings → Collabora Online**. Values saved in the GUI are stored in the File Browser database. CLI flags and environment variables are still supported and are used as defaults until the GUI settings are saved.

The same options can also be configured as CLI flags, config file entries, or environment variables:

| CLI flag | Environment variable | Purpose |
|---|---|---|
| `--collabora.enabled` | `FB_COLLABORA_ENABLED` | Enables the integration. |
| `--collabora.url` | `FB_COLLABORA_URL` | Public Collabora Online URL, for example `https://collabora.example.com`. |
| `--collabora.publicURL` | `FB_COLLABORA_PUBLIC_URL` | External/public File Browser base URL that Collabora can call back to, for example `https://files.example.com`. Include `baseURL` here if File Browser is mounted under a path. |
| `--collabora.internalURL` | `FB_COLLABORA_INTERNAL_URL` | Optional internal/LAN File Browser base URL used when File Browser is opened from LAN, for example `http://filebrowser.lan:8080`. |
| `--collabora.wopiSecret` | `FB_COLLABORA_WOPI_SECRET` | Secret for short-lived WOPI access tokens. If empty, the existing File Browser key is used. |
| `--collabora.tokenTTL` | `FB_COLLABORA_TOKEN_TTL` | WOPI token lifetime. Default: `2h`. |

Example:

```yaml
services:
  filebrowser:
    image: your-filebrowser-collabora-image:latest
    environment:
      FB_ADDRESS: 0.0.0.0
      FB_PORT: 8080
      FB_ROOT: /srv
      FB_COLLABORA_ENABLED: "true"
      FB_COLLABORA_URL: https://collabora.example.com
      FB_COLLABORA_PUBLIC_URL: https://files.example.com
      FB_COLLABORA_INTERNAL_URL: http://filebrowser.lan:8080
      FB_COLLABORA_WOPI_SECRET: change-this-long-random-secret
      FB_COLLABORA_TOKEN_TTL: 2h
    volumes:
      - ./filebrowser.db:/database/filebrowser.db
      - /mnt/data:/srv
```

## What was added

Backend:

- `/api/collabora/open?path=/file.docx`
- `/wopi/files/{id}` for `CheckFileInfo`, `LOCK`, `UNLOCK`, `REFRESH_LOCK`, `UNLOCK_AND_RELOCK`, and `GET_LOCK`
- `/wopi/files/{id}/contents` for `GetFile` and `PutFile`
- discovery lookup from `FB_COLLABORA_URL/hosting/discovery`
- short-lived JWT/HMAC WOPI tokens
- permission checks against the logged-in File Browser user
- CSP changes to allow the configured Collabora URL in frames and WebSocket connections

Frontend:

- `Open with Collabora` button for supported office files
- `/files/...?...office=true` iframe editor view
- Collabora API client
- Admin settings form for Collabora URL, public File Browser URL, WOPI secret, and token TTL

## Important Collabora-side requirement

Collabora must allow the public File Browser host as an accepted WOPI host. If Collabora logs show messages like `No acceptable WOPI hosts found`, add the File Browser hostname to Collabora's WOPI allow list / `aliasgroup` configuration.

## Current scope

This first version supports normal open/edit/save through WOPI locks and `PutFile`. Advanced WOPI operations such as `PutRelativeFile`, rename, delete, and co-authoring persistence beyond this single File Browser process are not implemented yet.

## Internal and external File Browser URLs

When both fields are configured, File Browser chooses the WOPI callback URL dynamically:

- if the browser opened File Browser through the internal/LAN URL, WOPI uses `internalURL`;
- if the browser opened File Browser through the external/public URL, WOPI uses `publicURL`;
- if the current request host is a private IP or `.home`/`.lan`/`.local` hostname, File Browser prefers `internalURL`;
- otherwise it falls back to `publicURL`.

For your mixed setup this allows both of these to work:

```text
http://filebrowser.lan:8080
https://files.example.com
```

Collabora must allow both WOPI hosts, for example:

```yaml
environment:
  aliasgroup1: 'https://files\.example\.com:443'
  aliasgroup2: 'https://files\.example\.com'
  aliasgroup3: 'http://filebrowser\.lan:8080'
```
