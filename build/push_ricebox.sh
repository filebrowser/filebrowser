#!/bin/sh

set -e

cd $(dirname $0)

COMMIT_SHA="$(git rev-parse --verify HEAD | cut -c1-8)"

eval `ssh-agent -s`
openssl aes-256-cbc -K $encrypted_9ca81b5594f5_key -iv $encrypted_9ca81b5594f5_iv -in ./deploy_key.enc -d | ssh-add -

git clone git@github.com:filebrowser/caddy caddy
cd caddy
cp ../../lib/rice-box.go assets/
sed -i 's/package lib/package assets/g' assets/rice-box.go
git checkout -b update-rice-box origin/master
git config --local user.name "Filebrowser Bot"
git config --local user.email "FilebrowserBot@users.noreply.github.com"
git commit -am "update rice-box $COMMIT_SHA"

if [ $(git tag | grep "$TRAVIS_TAG" | wc -l) -ne 0 ]; then
  git tag -d "$TRAVIS_TAG"
fi

git tag "$TRAVIS_TAG"

if [ "$(git ls-remote --heads origin update-rice-box)" != "" ]; then
  git push -u origin update-rice-box
else
  git push origin +update-rice-box
fi

if [ "$(git ls-remote --heads origin update-rice-box)" != "" ]; then
  git push origin "$TRAVIS_TAG"
else
  git push origin :"$TRAVIS_TAG"
  git push origin "$TRAVIS_TAG"
fi

