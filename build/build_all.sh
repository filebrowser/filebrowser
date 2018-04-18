#!/bin/sh

cd $(dirname $0)/..

if [ "$USE_DOCKER" != "" ]; then
  if [ -d "frontend/dist" ]; then
    rm -rf frontend/dist
  fi;

  if [ "$WDIR" = "" ]; then
    WDIR="/go/src/github.com/filebrowser/filebrowser"
  fi;

  $(command -v winpty) docker run -it \
    --name filebrowser-tmp \
    -v /$(pwd):/src:z \
    -w /${WDIR} \
    filebrowser/dev \
    sh -c "\
      cp -r //src/* /$WDIR && \
      cd build && \
      dos2unix build_assets.sh && \
      dos2unix build.sh && \
      ./build_assets.sh && \
      ./build.sh \
    "
  exitcode=$?

  if [ $exitcode -eq 0 ]; then
    for d in "dist/" "node_modules/"; do
      docker cp filebrowser-tmp:/$WDIR/frontend/$d frontend
    done
    for d in "vendor/" "rice-box.go" "filebrowser"; do
      docker cp filebrowser-tmp:/$WDIR/$d ./
    done
  fi
  docker rm -f filebrowser-tmp
else
  ./build/build_assets.sh
  ./build/build.sh
fi
