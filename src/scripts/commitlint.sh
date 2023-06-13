#!/usr/bin/env bash
set -e

if ! [ -x "$(command -v commitlint)" ]; then
  echo "commitlint is not installed. please run 'npm i -g commitlint'"
  exit 1
fi

for commit_hash in $(git log --pretty=format:%H origin/master..HEAD); do
  commitlint -f ${commit_hash}~1 -t ${commit_hash}
done
