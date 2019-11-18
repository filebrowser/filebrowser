#!/usr/bin/env sh

set -e

untracked="(untracked)"
REPO=$(cd $(dirname $0); pwd)
COMMIT_SHA="true"
ASSETS="false"
BINARY="false"
RELEASE="false"
SEMVER=""

debugInfo () {
  echo "Repo:           $REPO"
  echo "Build assets:   $ASSETS"
  echo "Build binary:   $BINARY"
  echo "Release:        $SEMVER"
}

buildAssets () {
  cd $REPO
  rm -rf frontend/dist
  rm -f http/rice-box.go

  cd $REPO/frontend

  if [ "$CI" = "true" ]; then
    npm ci
  else
    npm install
  fi

  npm run lint
  npm run build
}

buildBinary () {
  if ! [ -x "$(command -v rice)" ]; then
    go install github.com/GeertJohan/go.rice/rice
  fi

  cd $REPO/http
  rm -rf rice-box.go
  rice embed-go

  cd $REPO
  CMD_VARS=""
  if [ "$COMMIT_SHA" = "true" ]; then
    SHA=$(git rev-parse --short HEAD)
    CMD_VARS="-X github.com/filebrowser/filebrowser/v2/version.CommitSHA=$SHA"
  fi
  
  go build -a -o filebrowser -ldflags "-s -w"
}

release () {
  cd $REPO

  echo "RELEASE $#"

  echo "👀 Checking semver format"

  if [ $# -ne 1 ]; then
    echo "❌ This release script requires a single argument corresponding to the semver to be released. See semver.org"
    exit 1
  fi

  GREP="grep"
  if [ -x "$(command -v ggrep)" ]; then
    GREP="ggrep"
  fi

  semver=$(echo "$1" | $GREP -P '^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)')

  if [ $? -ne 0 ]; then
    echo "❌ Not valid semver format. See semver.org"
    exit 1
  fi

  echo "🧼  Tidying up go modules"
  go mod tidy

  echo "🐑 Creating a new commit for the new release"
  git commit --allow-empty -am "chore: version $semver"
  git tag "$1"
  git push
  git push --tags origin

  echo "📦 Done! $semver released."
}

usage() {
  echo "Usage: $0 [-a] [-c] [-b] [-r <string>]" 1>&2;
  exit 1;
}

DEBUG="false"

while :; do
  case $1 in
    -h|-\?|--help)
      usage
      exit
      ;;
    -a|--assets)
      ASSETS="true"
      ;;
    -c|--compile)
      BINARY="true"
      ;;
    -b|--build)
      ASSETS="true"
      BINARY="true"
      ;;
    -r|--release)
      RELEASE="true"
      if [ "$2" ]; then
        SEMVER=$2
        shift
      fi
      ;;
    --release=?*)
      RELEASE="true"
      SEMVER=${1#*=}
      ;;
    --release=)
      RELEASE="true"
      ;;
    -d|--debug)
      DEBUG="true"
      ;;
    --nosha)
      COMMIT_SHA="false"
      ;;
    --)
      shift
      break
      ;;
    -?*)
      usage
      ;;
    *)
    break
  esac
  shift
done

if [ "$DEBUG" = "true" ]; then
  debugInfo
fi

if [ "$ASSETS" = "true" ]; then
  buildAssets
fi

if [ "$BINARY" = "true" ]; then
  buildBinary
fi

if [ "$RELEASE" = "true" ]; then
  release $SEMVER
fi
