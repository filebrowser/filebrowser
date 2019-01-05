#!/bin/sh

set -e

if [ "$1" = 'run' ]; then
  if [ ! -f "/database.db" ]; then
    filebrowser -s /src
  fi

  exec filemanager --port 80
fi

exec filemanager --port 80 "$@"
