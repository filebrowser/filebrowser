#!/bin/sh

set -e

cd $(dirname $0)/..

if [ "$USE_DOCKER" != "" ]; then
  $(command -v winpty) docker run --rm -itv "/$(pwd)://src" -w "//src" filebrowser/dev sh -c "\
    go get -v ./... && \
    golangci-lint run -v"
else
  golangci-lint run -v
fi
