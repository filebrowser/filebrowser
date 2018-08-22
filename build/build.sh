#!/bin/sh

set -e

cd $(dirname $0)/../cli

go get -v ./...

if [ "$COMMIT_SHA" != "" ]; then
  echo "Set version to ($COMMIT_SHA)"
  sed -i.bak "s|(untracked)|($COMMIT_SHA)|g" ../lib/filebrowser.go
fi

echo "Build CLI"
CGO_ENABLED=0 go build -a -o filebrowser

if [ "$COMMIT_SHA" != "" ]; then
  echo "Reset version to (untracked)"
  sed -i "s|($COMMIT_SHA)|(untracked)|g" ../lib/filebrowser.go
fi
