ℹ INFO: **This project is not under active development ATM. A small group of developers keeps the project alive, but due to lack of time, we can't continue adding new features or doing deep changes. Please read [#532](https://github.com/filebrowser/filebrowser/issues/532) for more info!**

# filebrowser

[![Travis](https://img.shields.io/travis/com/filebrowser/filebrowser.svg?style=flat-square)](https://travis-ci.com/filebrowser/filebrowser)
[![Go Report Card](https://goreportcard.com/badge/github.com/filebrowser/filebrowser?style=flat-square)](https://goreportcard.com/report/github.com/filebrowser/filebrowser)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/filebrowser/filebrowser)
[![Version](https://img.shields.io/github/release/filebrowser/filebrowser.svg?style=flat-square)](https://github.com/filebrowser/filebrowser/releases/latest)
[![Chat IRC](https://img.shields.io/badge/freenode-%23filebrowser-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23filebrowser)

filebrowser provides a file managing interface within a specified directory and it can be used to upload, delete, preview, rename and edit your files. It allows the creation of multiple users and each user can have its own directory. It can be used as a standalone app or as a middleware.

# Table of contents

+ [Quick Start](#quick-start)
+ [Features](#features)
+ [Installation](#installation)
  - [One-step script](#one-step-script)
  - [Docker](#docker)
+ [Usage](#usage)
+ [Command line interface](#command-line-interface)
+ [Contributing](#contributing)

# Quick Start

The fastest way for beginners to start using File Browser is by following the instructions bellow. Although, there are [other ways](#installation) to install/use it.

- [Download File Browser](https://github.com/filebrowser/filebrowser/releases).
- Put the binary in your PATH.
- Run `filebrowser -s /path/to/your/files`.

Done! It will tell you the address in which File Browser is running. You only need to open it and use the following credentials (you should change them!):

- Username: ```admin```
- Password: ```admin```

Although this is the fastest way to run File Browser, we recommend you to take a look at its [usage](#usage) as it contains more information about what you can do.

# Features

# Installation

## One-step script

If you're running a Linux distribution or macOS, you can use our special script - made by [Kyle Frost](https://www.kylefrost.me/) - to download the latest version of File Browser and install it on `/usr/local/bin`.

```shell
curl -fsSL https://filebrowser.github.io/get.sh | bash
```

If you're on Windows, you can use PowerShell to install File Browser too. You should run the command as administrator because it needs perissions to add the executable to the PATH:

```shell
iwr -useb https://filebrowser.github.io/get.ps1 | iex
```

## Docker

# Usage

# Command line interface

# Contributing

The contributing guidelines can be found [here](https://github.com/filebrowser/community).
