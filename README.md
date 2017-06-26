# filemanager

[![Build](https://img.shields.io/travis/hacdias/filemanager.svg?style=flat-square)](https://travis-ci.org/hacdias/filemanager)
[![Go Report Card](https://goreportcard.com/badge/github.com/hacdias/filemanager?style=flat-square)](https://goreportcard.com/report/hacdias/filemanager)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/hacdias/filemanager)


## About Search

FileManager allows you to search through your files and it has some options. By default, your search will be something like this:

```
this are keywords
```

If you search for that it will look at every file that contains "this", "are" or "keywords" on their name. If you want to search for an exact term, you should surround your search by double quotes:

```
"this is the name"
```

That will search for any file that contains "this is the name" on its name. It won't search for each separated term this time.

By default, every search will be case sensitive. Although, you can make a case insensitive search by adding `case:insensitive` to the search terms, like this:

```
this are keywords case:insensitive
```