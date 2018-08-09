#!/bin/sh

set -e

cd $(dirname $0)/..

dep ensure -vendor-only

if [ "$COMMIT_SHA" != "" ]; then
  echo "Set version to ($COMMIT_SHA)"
  sed -i.bak "s|(untracked)|($COMMIT_SHA)|g" filebrowser.go
fi

echo "Build cmd/filebrowser"
cd cmd/filebrowser
CGO_ENABLED=0 go build -a
cd ../..
cp cmd/filebrowser/filebrowser ./

if [ "$COMMIT_SHA" != "" ]; then
  echo "Reset version to (untracked)"
  sed -i "s|($COMMIT_SHA)|(untracked)|g" filebrowser.go
fi
