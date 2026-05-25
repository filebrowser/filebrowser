# File Browser ConvertX "Convert to..." action

This build adds a server-side ConvertX integration to File Browser.

## GUI settings

Configure ConvertX from **Admin → Settings → Global Settings → ConvertX**:

- Enable ConvertX integration
- ConvertX URL, for example `http://convertx.lan:3000`
- ConvertX API key, matching `CONVERTX_API_TOKEN` in ConvertX
- ConvertX request timeout
- Test connection

The API key is stored in File Browser settings and is only used by the File Browser backend. It is not sent to the browser.

The connection test requires the ConvertX File Browser integration API. A normal ConvertX browser page is not enough for this feature.

## User action

When ConvertX is enabled and a user selects one regular file, File Browser shows **Convert to...** in:

- the desktop toolbar
- the mobile selected-file action row
- the right-click context menu

The user selects a target format/converter. File Browser then:

1. Reads the selected file from the user's File Browser filesystem.
2. Uploads it server-side to ConvertX `/api/convert` using the configured bearer token.
3. Polls ConvertX `/api/jobs/:jobId`.
4. Downloads the converted output from ConvertX.
5. Saves the converted file back into the same File Browser directory.

By default, existing files are not overwritten. File Browser automatically adds a numeric suffix when needed, for example `document(1).pdf`.

## File Browser endpoints

```http
GET  /api/convertx/targets?from=docx
POST /api/convertx/convert
GET  /api/convertx/jobs/:jobId
```

`POST /api/convertx/convert` accepts:

```json
{
  "path": "/folder/document.docx",
  "convertTo": "pdf",
  "converter": "libreoffice",
  "rename": true,
  "overwrite": false
}
```

The operation is asynchronous by default and returns a File Browser job id. The frontend polls `/api/convertx/jobs/:jobId` until the result is saved.

## Required ConvertX API

This requires the ConvertX API archive that adds:

```http
GET  /api/health
GET  /api/conversions?from=docx
POST /api/convert
GET  /api/jobs/:jobId
GET  /api/jobs/:jobId/files/:fileName
GET  /api/jobs/:jobId/file
```

Set this in ConvertX:

```bash
CONVERTX_API_TOKEN=change-this-secret
```
