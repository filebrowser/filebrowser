#!/bin/sh

set -e

PORT=${FB_PORT:-$(cat /tmp/FB_CONFIG | sh /JSON.sh | grep '\["port"\]' | awk '{print $2}')}
ADDRESS=${FB_ADDRESS:-$(cat /tmp/FB_CONFIG | sh /JSON.sh | grep '\["address"\]' | awk '{print $2}' | sed 's/"//g')}
ADDRESS=${ADDRESS:-localhost}

wget -q --spider http://$ADDRESS:$PORT/health || exit 1