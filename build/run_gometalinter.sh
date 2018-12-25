#!/bin/sh

set -e

cd $(dirname $0)/..

dolint='gometalinter --exclude="rice-box.go" --exclude="vendor" --deadline=300s ./...'

if [ "$USE_DOCKER" != "" ]; then
  $(command -v winpty) docker run --rm -itv /$(pwd)://go/src/github.com/filebrowser/filebrowser -w //go/src/github.com/filebrowser/filebrowser filebrowser/dev:mod sh -c "\
    GO111MODULE=on go get -v ./... && \
    GO111MODULE=on go mod vendor && \
    GO111MODULE=off $dolint"
else
  $dolint
fi
