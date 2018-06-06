#!/bin/sh

cd $(dirname $0)/..

if [ "$USE_DOCKER" != "" ]; then
  if [ -d "frontend/dist" ]; then
    rm -rf frontend/dist
  fi;
  WORKDIR="/go/src/github.com/filebrowser/filebrowser"

  $(command -v winpty) docker run -it \
    --name filebrowser-tmp \
    -v /$(pwd):/src:z \
    -w /${WORKDIR} \
    filebrowser/dev \
    sh -c "\
      cp -r //src/* /$WORKDIR && \
      cd build && \
      dos2unix build_assets.sh && \
      dos2unix build.sh && \
      ./build_assets.sh && \
      ./build.sh \
    "
  exitcode=$?

  if [ $exitcode -eq 0 ]; then
    for d in "dist/" "node_modules/"; do
      docker cp filebrowser-tmp:/$WORKDIR/frontend/$d frontend
    done
    for d in "vendor/" "rice-box.go" "filebrowser"; do
      docker cp filebrowser-tmp:/$WORKDIR/$d ./
    done
  fi
  docker rm -f filebrowser-tmp
else
  ./build/build_assets.sh
  ./build/build.sh
fi
