#!/bin/sh

set -e

PORT=${FB_PORT:-$(jq -r .port /config/settings.json)}
ADDRESS=${FB_ADDRESS:-$(jq -r .address /config/settings.json)}
ADDRESS=${ADDRESS:-localhost}

wget -q --spider http://$ADDRESS:$PORT/health || exit 1
