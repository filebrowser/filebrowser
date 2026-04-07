#!/bin/sh

set -e

# Ensure configuration exists
if [ ! -f "/config/settings.json" ]; then
  cp -a /defaults/settings.json /config/settings.json
fi

# Extract config file path from arguments
config_file=""
next_is_config=0
for arg in "$@"; do
  if [ "$next_is_config" -eq 1 ]; then
    config_file="$arg"
    break
  fi
  case "$arg" in
    -c|--config)
      next_is_config=1
      ;;
    -c=*|--config=*)
      config_file="${arg#*=}"
      break
      ;;
  esac
done

# If no config argument is provided, set the default and add it to the args                                                                 
if [ -z "$config_file" ]; then 
  config_file="/config/settings.json"                                                                                                                                                                                                 
  set -- --config=/config/settings.json "$@"                                                                                                       
fi                                                                                                                                                                                                                                                                                                                                                             

exec filebrowser "$@"
