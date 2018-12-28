#!/bin/sh

set -e

cd $(dirname $0)/..

# Clean the dist folder and build the assets
cd frontend
if [ -d "dist" ]; then
  rm -rf dist/*
fi;
yarn install
yarn build
cd ..

# Install rice tool if not present
if ! [ -x "$(command -v rice)" ]; then
  go get github.com/GeertJohan/go.rice/rice
fi

# Embed the assets using rice
cd lib
rice embed-go
