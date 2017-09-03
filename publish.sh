#!/bin/bash

echo "Releasing version $1"
sed -i "s|(untracked version)|$1|g" filemanager.go

echo "Building assets"
./build.sh

echo "Commiting..."
git add -A
git commit -m "Version $1"
git push

echo "Creating the tag..."
git tag "v$1"
git push --tags

echo "Commiting untracked version notice..."
sed -i "s|$1|(untracked version)|g" filemanager.go
git add -A
git commit -m "untracked version `date`"
git push

echo "Done!"
