#!/bin/sh

set -e

cd $(dirname $0)/..

mkdir -p tmp-dev
cd tmp-dev
cp ../dockerfiles/dev Dockerfile
docker build -t filebrowser/dev .
cd ..
rm -rf tmp-dev
