#!/bin/sh

cd $(dirname $0)

docker pull golang:alpine

$(command -v winpty) docker run --rm -it \
  -v /$(pwd)://src \
  -w //src \
  golang:alpine \
  sh -c '\
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
    sed -i -e "s/v[0-9]\.[0-9]/edge/g" /etc/apk/repositories  && \
    apk add -U --no-cache yarn git  && \
    go get github.com/GeertJohan/go.rice/rice && \
    ./build.sh \
  '
