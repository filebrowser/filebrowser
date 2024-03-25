#!/bin/sh
PORT=${FB_PORT:-$(jq .port /.filebrowser.json)}
ADDRESS=${FB_ADDRESS:-$(jq .address /.filebrowser.json)}
ADDRESS=${ADDRESS:-localhost}
curl -f http://$ADDRESS:$PORT/health || exit 1
