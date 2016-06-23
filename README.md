# filemanager - a caddy plugin

[![Build](https://img.shields.io/travis/hacdias/caddy-filemanager.svg?style=flat-square)](https://travis-ci.org/hacdias/caddy-filemanager)
[![community](https://img.shields.io/badge/community-forum-ff69b4.svg?style=flat-square)](https://forum.caddyserver.com)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/hacdias/caddy-filemanager)

This package is a plugin for Caddy server that provides an online file manager (based on browse middleware) that is able to: rename files, delete files and upload files. Some new features that can be implemented in the future can be seen at [issues](https://github.com/hacdias/caddy-filemanager/issues).

```
filemanager {
  show    path
  on      url
  styles  filepath
}
```
