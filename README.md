# filemanager - a caddy plugin

[![Build](https://img.shields.io/travis/hacdias/caddy-filemanager.svg?style=flat-square)](https://travis-ci.org/hacdias/caddy-filemanager)
[![community](https://img.shields.io/badge/community-forum-ff69b4.svg?style=flat-square)](https://forum.caddyserver.com)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/hacdias/caddy-filemanager)
[![Go Report Card](https://goreportcard.com/badge/github.com/hacdias/caddy-filemanager?style=flat-square)](https://goreportcard.com/report/hacdias/caddy-filemanager)

This package is a plugin for Caddy server that provides an online file manager (based on browse middleware) that is able to: rename files, delete files and upload files. Some new features that can be implemented in the future can be seen at [issues](https://github.com/hacdias/caddy-filemanager/issues).

### Syntax

```
filemanager url {
  show              path
  webdav            [path]
  styles            filepath
  frontmatter       type
  allow_new         [true|false]
  allow_edit        [true|false]
  allow_commands    [true|false]
  allow_command     command
  block_command     command
  before_save       command
  after_save        command
  allow             [path|dotfiles]
  allow_r           path regex
  block             [path|dotfiles]
  block_r           path regex
}
```


## NOTE FOR DEVELOPERS

You need to run `go generate` on `$GOPATH/src/github.com/hacdias/caddy-filemanager` before building any binary. Otherwise, you will receive an `undefined: Asset` error.
