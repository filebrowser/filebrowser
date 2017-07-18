# filemanager

[![Build](https://img.shields.io/travis/hacdias/filemanager.svg?style=flat-square)](https://travis-ci.org/hacdias/filemanager)
[![Go Report Card](https://goreportcard.com/badge/github.com/hacdias/filemanager?style=flat-square)](https://goreportcard.com/report/hacdias/filemanager)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/hacdias/filemanager)

filemanager provides a file managing interface within a specified directory and it can be used to upload, delete, preview, rename and edit your files. It allows the creation of multiple users and each user can have its own directory. It can be used as a standalone app or as a middleware.

# Features

Easy login system.

![Login Page](https://user-images.githubusercontent.com/5447088/28327862-931bb42a-6bdc-11e7-8157-93fa54945f3c.png)

Listings of your files, available in two styles: mosaic and list. You can delete, move, rename, upload and create new files, as well as directories. Single files can be downloaded directly, and multiple files as *.zip*, *.tar*, *.tar.gz*, *.tar.bz2* or *.tar.xz*.

![Mosaic Listing](https://user-images.githubusercontent.com/5447088/28327863-931fe414-6bdc-11e7-91fb-49b5f15a829f.png)

File Manager editor is powered by [Codemirror](https://codemirror.net/) and if you're working with markdown files with metadata, both parts will be separated from each other so you can focus on the content.

![Markdown Editor](https://user-images.githubusercontent.com/5447088/28327865-933abc3a-6bdc-11e7-8f80-f206cf5cdf0b.png)

On the settings page, a regular user can set its own custom CSS to personalize the experience and change its password. For admins, they can manage the permissions of each user, set commands which can be executed when certain events are triggered (such as before saving and after saving) and change plugin's settings.

![Settings](https://user-images.githubusercontent.com/5447088/28327864-9325d716-6bdc-11e7-9de2-78953e6efdbe.png)

We also allow the users to search in the directories and execute commands if allowed.


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
