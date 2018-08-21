#!/bin/sh

set -e

cd $(dirname $0)/..

dolint='gometalinter --exclude="rice-box.go" --exclude="vendor" --deadline=300s ./...'

if [ "$USE_DOCKER" != "" ]; then
  docker run --rm -itv $(pwd):/src filebrowser/dev sh -c "\
    cp -r /src/. ./ && cd cli && go get -v ./... && \
    CGO_ENABLED=0 $dolint"
else
  $dolint
fi
