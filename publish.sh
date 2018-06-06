#!/bin/bash
set -e

cd $(dirname $0)

echo "Building assets"
./build/build_assets.sh

echo "Updating version number to $1..."
sed -i "s|(untracked)|$1|g" filebrowser.go
git add -A
git commit -m "chore: version $1"
git tag "v$1"
git push
git push --tags

echo "Commiting untracked version notice..."
sed -i "s|$1|(untracked)|g" filebrowser.go
git add -A
git commit -m "chore: setting untracked version [ci skip]"
git push

echo "Done!"
