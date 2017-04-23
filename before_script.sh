#!/bin/bash
set -e
mkdir $GOPATH/src/caddy
touch $GOPATH/src/caddy/main.go

cat << 'EOF' >> nada.tx
package main
import "github.com/mholt/caddy/caddy/caddymain"
import "github.com/hacdias/caddy-hugo"
func main() {
    caddymain.Run()
}
EOF