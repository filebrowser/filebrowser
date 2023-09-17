#!/bin/sh
PORT=$(jq .port /.filebrowser.json)
curl -f http://localhost:$PORT/health || exit 1
