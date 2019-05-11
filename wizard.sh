#!/usr/bin/env sh

set -e

untracked="(untracked)"
REPO=$(cd $(dirname $0); pwd)
COMMIT_SHA=$(git rev-parse --short HEAD)
ASSETS="false"
BINARY="false"
RELEASE=""

debugInfo () {
  echo "Repo:           $REPO"
  echo "Build assets:   $ASSETS"
  echo "Build binary:   $BINARY"
  echo "Release:        $RELEASE"
}

updateVersion () {
  from=$1
  to=$2

  echo "🎁 Updating version from \"$from\" to \"$to\""
  sed -i.bak "s|$from|$to|g" $REPO/version/version.go
}

buildAssets () {
  cd $REPO
  rm -rf frontend/dist
  rm -f http/rice-box.go

  cd $REPO/frontend
  npm install
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
  updateVersion $untracked "($COMMIT_SHA)"
  go build -a -o filebrowser
  updateVersion "($COMMIT_SHA)" $untracked
}

release () {
  cd $REPO

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

  echo "🐑 Fetching 'master' on 'frontend' and creating new tag"

  cd frontend
  git fetch --all
  git checkout master
  git tag $semver
  git push origin $semver

  cd ..

  echo "🐑 Updating submodule ref to $semver"
  updateVersion $untracked $1
  git commit -am "chore: version $semver"
  git tag "$1"
  git push
  git push origin $semver

  echo "🐑 Commiting untracked version notice..."
  updateVersion $1 $untracked
  git commit -am "chore: setting untracked version [ci skip]"
  git push

  echo "📦 Done! $semver released."
}

usage() {
  echo "Usage: $0 [-a] [-c] [-b] [-r <string>]" 1>&2;
  exit 1;
}

DEBUG="false"

while getopts "bacr:d" o; do
  case "${o}" in
    b)
      ASSETS="true"
      BINARY="true"
      ;;
    a)
      ASSETS="true"
      ;;
    c)
      BINARY="true"
      ;;
    r)
      RELEASE=${OPTARG}
      ;;
    d)
      DEBUG="true"
      ;;
    *)
      usage
      ;;
  esac
done
shift $((OPTIND-1))

if [ "$DEBUG" = "true" ]; then
  debugInfo
fi

if [ "$ASSETS" = "true" ]; then
  buildAssets
fi

if [ "$BINARY" = "true" ]; then
  buildBinary
fi

if [ "$RELEASE" != "" ]; then
  release $RELEASE
fi
