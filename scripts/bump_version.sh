#!/usr/bin/env bash
set -e

if ! [ -x "$(command -v npx)" ]; then
  echo "Node.js is require, exiting..."
  exit 1
fi

npx standard-version --dry-run --skip
read -p "Continue (y/n)? " -n 1 -r
echo ;
if [[ $REPLY =~ ^[Yy]$ ]]; then
	npx standard-version -s ;
fi