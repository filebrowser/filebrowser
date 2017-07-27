![Preview](https://user-images.githubusercontent.com/5447088/28537288-39be4288-70a2-11e7-8ce9-0813d59f46b7.gif)

# filemanager

[![Build](https://img.shields.io/travis/hacdias/filemanager.svg?style=flat-square)](https://travis-ci.org/hacdias/filemanager)
[![Go Report Card](https://goreportcard.com/badge/github.com/hacdias/filemanager?style=flat-square)](https://goreportcard.com/report/hacdias/filemanager)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/hacdias/filemanager)

filemanager provides a file managing interface within a specified directory and it can be used to upload, delete, preview, rename and edit your files. It allows the creation of multiple users and each user can have its own directory. It can be used as a standalone app or as a middleware.

# Table of contents

+ [Getting started](#getting-started)
  - [Caddy](#caddy)
  - [Standalone](#standalone)
+ [Features](#features)
  - [Users](#users)
  - [Search](#search)
+ [Contributing](#contributing)
+ [Donate](#donate)

# Getting started

This is a library that can be used on your own applications as a middleware (see the [documentation](http://godoc.org/github.com/hacdias/filemanager)), as a plugin to Caddy web server or as a standalone app.

Once you have everything deployed, the default credentials to login to the filemanager are:

**Username:** `admin`
**Password:** `admin`

## Caddy

The easiest way to get started is using this with Caddy web server. You just need to download Caddy from its [official website](https://caddyserver.com/download) with `http.filemanager` plugin enabled. For more information about the plugin itself, please refer to its [documentation](https://caddyserver.com/docs/http.filemanager).

## Standalone

You can use filemanager as a standalone executable. You just need to download it from the [releases page](https://github.com/hacdias/filemanager/releases), where you can find multiple releases.

You can either use flags or a JSON configuration file, which should have the following appearance:

```json
{
  "port": 80,
  "address": "127.0.0.1",
  "database": "/path/to/database.db",
  "scope": "/path/to/my/files",
  "allowCommands": true,
  "allowEdit": true,
  "allowNew": true,
  "commands": [
    "git",
    "svn"
  ]
}
```

The `scope`, `allowCommands`, `allowEdit`, `allowNew` and `commands` options are the defaults for new users. To set a configuration file, you will need to pass the path with a flag, like this: `filemanager --config=/path/to/config.json`.

Otherwise, you may not want to use a configuration file, which can be done using the following flags:

```
-address string
      Address to listen to (default is all of them)
-allow-commands
      Default allow commands option (default true)
-allow-edit
      Default allow edit option (default true)
-allow-new
      Default allow new option (default true)
-commands string
      Space separated commands available for new users (default "git svn hg")
-database string
      Database path (default "./filemanager.db")
-port string
      HTTP Port (default is random)
-scope string
      Default scope for new users (default ".")
```

## Docker

(TODO)

# Features

Easy login system.

![Login Page](https://user-images.githubusercontent.com/5447088/28432382-975493dc-6d7f-11e7-9190-23f8037159dc.jpg)

Listings of your files, available in two styles: mosaic and list. You can delete, move, rename, upload and create new files, as well as directories. Single files can be downloaded directly, and multiple files as *.zip*, *.tar*, *.tar.gz*, *.tar.bz2* or *.tar.xz*.

![Mosaic Listing](https://user-images.githubusercontent.com/5447088/28432384-9771bb4c-6d7f-11e7-8564-3a9bd6a3ce3a.jpg)

File Manager editor is powered by [Codemirror](https://codemirror.net/) and if you're working with markdown files with metadata, both parts will be separated from each other so you can focus on the content.

![Markdown Editor](https://user-images.githubusercontent.com/5447088/28432383-9756fdac-6d7f-11e7-8e58-fec49470d15f.jpg)

On the settings page, a regular user can set its own custom CSS to personalize the experience and change its password. For admins, they can manage the permissions of each user, set commands which can be executed when certain events are triggered (such as before saving and after saving) and change plugin's settings.

![Settings](https://user-images.githubusercontent.com/5447088/28432385-9776ec66-6d7f-11e7-90a5-891bacd4d02f.jpg)

We also allow the users to search in the directories and execute commands if allowed.

## Users

We support multiple users and each user can have its own scope and custom stylesheet. The administrator is able to choose which permissions should be given to the users, as well as the commands they can execute. Each user also have a set of rules, in which he can be prevented or allowed to access some directories (regular expressions included!).

![Users](https://user-images.githubusercontent.com/5447088/28432386-977f388a-6d7f-11e7-9006-87d16f05f1f8.jpg)

## Search

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

# Contributing

If you want to contribute or want to build the code from source, you will need to have the most recent version of Go and, if you want to change the static assets (JS, CSS, ...), Node.js installed on your computer. To start developing, you just need to do the following:

1. `go get github.com/hacdias/filemanager`
2. `cd $GOPATH/src/github.com/hacdias/filemanager`
3. `npm install`
4. `npm start dev` - regenerates the static assets automatically
5. `go install gihthub.com/hacdias/filemanager/cmd/filemanager`
6. Execute `$GOPATH/bin/filemanager`

The steps 3 and 4 are only required **if you want to develop the front-end**. Otherwise, you can ignore them. Before pulling, if you made any change on assets folder, you must run the `build.sh` script on the root of this repository.

If you are using this as a Caddy plugin, you should use its [official instructions for plugins](https://github.com/mholt/caddy/wiki/Extending-Caddy#2-plug-in-your-plugin) and import `github.com/hacdias/filemanager/caddy/filemanager`.

# Donate

Enjoying this project? You can [donate to its creator](https://henriquedias.com/donate/). He will appreciate.
