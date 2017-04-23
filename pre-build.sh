#!/bin/sh

go get github.com/jteeuwen/go-bindata/go-bindata

go-bindata -pkg assets -prefix "_embed" \
  -o assets/binary.go -ignore "^.*theme-([^g]|g[^i]|gi[^t]|git[^h]|gith[^u]|githu[^b]).*\.js$"  \
  _embed/templates/... _embed/public/js/... _embed/public/css/... _embed/public/ace/src-min/... \

git add -A
