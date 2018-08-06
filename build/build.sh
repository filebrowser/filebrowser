#!/bin/sh

set -e

cd $(dirname $0)/..

dep ensure -vendor-only

cd cmd/filebrowser
CGO_ENABLED=0 go build -a
cd ../..
cp cmd/filebrowser/filebrowser ./
