#!/usr/bin/env bash
set -e

if ! [ -x "$(command -v npx)" ]; then
  echo "Node.js is require, exiting..."
  exit 1
fi

for commit_hash in $(git log --pretty=format:%H origin/master..HEAD); do
  npx commitlint -f ${commit_hash}~1 -t ${commit_hash}
done
