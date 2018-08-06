#!/bin/sh

set -e

cd $(dirname $0)/..

cp dockerfiles/filebrowser Dockerfile
docker build -t filebrowser/filebrowser .
rm -f Dockerfile
