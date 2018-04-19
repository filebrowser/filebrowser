#!/bin/sh

set -e

cd $(dirname $0)

WORKDIR="/go/src/github.com/filebrowser/filebrowser"

$(command -v winpty) docker run --rm -it \
  -v /$(pwd):/${WORKDIR} \
  -w /${WORKDIR} \
  filebrowser/filebrowser:dev \
  sh -c '\
    dos2unix build.sh && \
    ./build.sh \
  '
