#!/bin/sh
PORT=${FB_PORT:-$(jq -r .port filebrowser.json)}
ADDRESS=${FB_ADDRESS:-$(jq -r .address filebrowser.json)}
ADDRESS=${ADDRESS:-localhost}
VAR=$(jq -r .cert filebrowser.json)
if [ -z "${VAR}" ]; then
    curl -f http://$ADDRESS:$PORT/health || exit 1
else
    curl -f https://$ADDRESS:$PORT/health || exit 1
fi