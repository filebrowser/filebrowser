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

  $(command -v winpty) docker run -it \
    --name filebrowser-tmp \
    -v /$(pwd):/src:z \
    -w //src \
    -e COMMIT_SHA=$COMMIT_SHA \
    filebrowser/dev:mod \
    sh -c "\
      cd build && \
      dos2unix build_assets.sh && \
      dos2unix build.sh && \
      ./build_assets.sh && \
      ./build.sh \
    "
  exitcode=$?

  if [ $exitcode -eq 0 ]; then
    for d in "dist/" "node_modules/"; do
      docker cp filebrowser-tmp://src/frontend/$d frontend
    done
    docker cp filebrowser-tmp://src/cli/filebrowser ./filebrowser
    docker cp filebrowser-tmp://src/lib/rice-box.go ./lib/rice-box.go
  else
    echo "BUILD FAILED!"
  fi
  docker rm -f filebrowser-tmp
else
  set -e
  ./build/build_assets.sh
  ./build/build.sh
fi
