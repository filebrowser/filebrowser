#!/bin/sh

set -e

# Ensure configuration exists
if [ ! -f "/config/settings.json" ]; then
  cp -a /defaults/settings.json /config/settings.json
fi

# Deal with the case where user does not provide a config argument
has_config_arg=0
for arg in "$@"; do
  case "$arg" in
  --config|--config=*|-c|-c=*)
    has_config_arg=1
    break
    ;;
  esac
done

if [ "$has_config_arg" -eq 0 ]; then
  set -- --config=/config/settings.json "$@"
fi

exec filebrowser "$@"