#!/bin/bash

echo "Building assets"
./build.sh

echo "Updating version number to $1..."
sed -i "s|(untracked)|$1|g" filemanager.go

echo "Commiting..."
git add -A
git commit -m "Version $1"
git push

echo "Creating the tag..."
git tag "v$1"
git push --tags

echo "Commiting untracked version notice..."
sed -i "s|$1|(untracked)|g" filemanager.go
git add -A
git commit -m "[ci skip] auto: setting untracked version"
git push

echo "Done!"
