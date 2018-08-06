#!/bin/bash

set -e

cd $(dirname $0)/..

echo "> Checking semver format"

if [ $# -ne 1 ]; then
  echo "This release script requires a single argument corresponding to the semver to be released. See semver.org"
  exit 1
fi

semver=$(grep -P '^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)' <<< "$1")

if [ $? -ne 0 ]; then
  echo "Not valid semver format. See semver.org"
  exit 1
fi

echo "> Checking matching $semver in frontend submodule"

cd frontend
git fetch --all

if [ $(git tag | grep "$semver" | wc -l) -eq 0 ]; then
  echo "Tag $semver does not exist in submodule 'frontend'. Tag it and run this script again."
  exit 1
fi

git rev-parse --verify --quiet release
if [ $? -ne 0 ]; then
  git checkout -b release "$semver"
else
  git checkout release
  git reset --hard "$semver"
fi

cd ..

echo "> Updating submodule ref to $semver"

sed -i "s|(untracked)|$1|g" filebrowser.go
git commit -am "chore: version $semver"
git tag "$1"
git push
git push --tags

echo "> Commiting untracked version notice..."

sed -i "s|$1|(untracked)|g" filebrowser.go
git commit -am "chore: setting untracked version [ci skip]"
git push

echo "> Done!"
