#!/bin/bash
set -e

# Install rice tool if not present
if ! [ -x "$(command -v rice)" ]; then
  go get github.com/GeertJohan/go.rice/rice
fi

# Clean the dist folder and build the assets
rm -rf assets/dist
npm run build

# Embed the assets using rice
rice embed-go
