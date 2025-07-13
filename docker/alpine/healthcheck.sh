#!/bin/sh

set -e

CONFIG_FILE="/tmp/FB_CONFIG"
if [ ! -f "/tmp/FB_CONFIG" ]; then
  CONFIG_FILE="/config/settings.json"
fi

PORT=${FB_PORT:-$(cat $CONFIG_FILE | sh /JSON.sh | grep '\["port"\]' | awk '{print $2}')}
ADDRESS=${FB_ADDRESS:-$(cat $CONFIG_FILE | sh /JSON.sh | grep '\["address"\]' | awk '{print $2}' | sed 's/"//g')}
ADDRESS=${ADDRESS:-localhost}

wget -q --spider http://$ADDRESS:$PORT/health || exit 1