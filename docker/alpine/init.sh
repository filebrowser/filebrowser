#!/bin/sh

set -e

# Ensure configuration exists
if [ ! -f "/config/settings.json" ]; then
  cp -a /defaults/settings.json /config/settings.json
fi

exec "$@"
