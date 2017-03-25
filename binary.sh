#!/bin/sh

go get github.com/jteeuwen/go-bindata/go-bindata
go get github.com/bountylabs/gitversion
go-bindata -pkg assets -prefix "_embed" -o assets/binary.go -ignore=.*ace.* _embed/... _embed/public/js/vendor/ace/src-min/...
gitversion -s -o page/version.go -p page
