# hugo - a caddy plugin

[![Build](https://img.shields.io/travis/hacdias/caddy-hugo.svg?style=flat-square)](https://travis-ci.org/hacdias/caddy-hugo)
[![community](https://img.shields.io/badge/community-forum-ff69b4.svg?style=flat-square)](https://forum.caddyserver.com)
[![Documentation](https://img.shields.io/badge/caddy-doc-F06292.svg?style=flat-square)](https://caddyserver.com/docs/hugo)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/hacdias/caddy-hugo)

[Hugo](http://gohugo.io/) is an easy to use and fast command line static website generator, while [Caddy](http://caddyserver.com) is a lightweight, fast, general-purpose, cross-platform HTTP/2 web server with automatic HTTPS. This extension is able to bring a web interface to Caddy to manage Hugo generated websites. This plugin provides you an web interface to manage your websites made with Hugo.

**If you're not developer go to the [documentation](https://caddyserver.com/docs/hugo)**.

## Build from source

Requirements:

+ [Go 1.6 or higher][1]
+ [caddydev][2]
+ [go-bindata][3]
+ [Node.js w/ npm][4] (optional)

Instructions:

1. ```go get github.com/hacdias/caddy-hugo``` (ignore the error, see step 3)
2. ```cd $GOPATH/github.com/hacdias/caddy-hugo```
  1. If you want to modify the CSS/JS:
  2. Change the third comment to  ```//go:generate go-bindata -debug -pkg assets -o assets/assets.go templates/ assets/css/ assets/js/ assets/fonts/```
  3. ```npm install```
  4. ```grunt watch```
3. ```go generate```
4. ```cd $YOUR_WEBSITE_PATH```
5. ```caddydev --source $GOPATH/src/github.com/hacdias/caddy-hugo hugo```
6. Go to ```http://domain:port```

[1]: https://golang.org/dl/
[2]: https://github.com/caddyserver/caddydev
[3]: https://github.com/jteeuwen/go-bindata
[4]: https://nodejs.org
