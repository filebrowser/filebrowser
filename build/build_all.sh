#!/bin/sh

cd $(dirname $0)/..

if [ -d lib/"rice-box.go" ]; then
  rm -rf lib/rice-box.go
fi

if [ "$USE_DOCKER" != "" ]; then
  if [ -d "frontend/dist" ]; then
    rm -rf frontend/dist
  fi;

  if [ "$(command -v git)" != "" ]; then
    COMMIT_SHA="$(git rev-parse HEAD | cut -c1-8)"
  else
    COMMIT_SHA="untracked"
  fi

  $(command -v winpty) docker run --rm -it \
    -v /$(pwd):/src:z \
    -w //src \
    -e COMMIT_SHA=$COMMIT_SHA \
    filebrowser/dev \
    sh -c "\
      cd build && \
      dos2unix build_assets.sh && \
      dos2unix build.sh && \
      ./build_assets.sh && \
      ./build.sh \
    "
else
  set -e
  ./build/build_assets.sh
  ./build/build.sh
fi
