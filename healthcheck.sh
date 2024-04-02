#!/bin/sh
PORT=${FB_PORT:-$(jq -r .port /.filebrowser.json)}
ADDRESS=${FB_ADDRESS:-$(jq -r .address /.filebrowser.json)}
ADDRESS=${ADDRESS:-localhost}
curl -f http://$ADDRESS:$PORT/health || exit 1
