# filemanager - a caddy plugin

[![Build](https://img.shields.io/travis/hacdias/caddy-filemanager.svg?style=flat-square)](https://travis-ci.org/hacdias/caddy-filemanager)
[![community](https://img.shields.io/badge/community-forum-ff69b4.svg?style=flat-square)](https://forum.caddyserver.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/hacdias/caddy-filemanager?style=flat-square)](https://goreportcard.com/report/hacdias/caddy-filemanager)

filemanager provides WebDAV features and a file managing interface within the specified directory and it can be used to upload, delete, preview, rename and edit your files within that directory.

It is extremely important for security reasons to cover the path of filemanager with some kind of authentication. You can use, for example, [`basicauth`](https://caddyserver.com/docs/basicauth) directive.

Note that if you are handling large files you might run into troubles due to the defaults of [`timeouts`](https://caddyserver.com/docs/timeouts) plugin. Check its [documentation](https://caddyserver.com/docs/timeouts) to learn more about that plugin. 

## Get Started

To start using this plugin you just need to go to the [download Caddy page](https://caddyserver.com/download) and choose `filemanager` in the directives section. For further information on how Caddy works refer to [its documentation](https://caddyserver.com/docs).

If you want to build it from source, consult our [developers section](#developers).

## Syntax

```
filemanager [baseurl] {
    show           directory
    webdav         [path]
    styles         filepath
    allow_new      [true|false]
    allow_edit     [true|false]
    allow_commands [true|false]
    allow_command  command
    block_command  command
    before_save    command
    after_save     command
    allow          [url|dotfiles]
    allow_r        regex
    block          [url|dotfiles]
    block_r        regex
}
```

All of the options above are optional.

+ **baseurl** is the URL where you will access the File Manager interface. Defaults to `/`.
+ **show** is the path, relative or absolute, to the directory you want to browse in. Defaults to `./`.
+ **webdav** is the path that will be appended to baseurl in which the [WebDAV](https://en.wikipedia.org/wiki/WebDAV) will be accessible. Defaults to `/webdav`.
+ **styles** is the relative or absolute path to the stylesheet file. This file doesn't need to be accessible from the web.
+ **allow_new** is the permission to create new files and directories. Defaults to `true`.
+ **allow_edit** is the permission to edit, rename and delete files or directories. Defaults to `true`.
+ **allow_commands** is the permission to execute commands. Defaults to `true`.
+ **allow_command** and **block_command** gives, or denies, permission to execute a certain command through the admin interface. By default `git`, `svn` and `hg` are enabled.
+ **before_save** and **after_save** allow you to set a custom command to be executed before saving and after saving a file. The placeholder `{path}` can be used and it will be replaced by the file path.
+ **allow** and **block** can be used to allow or deny the access to specific files or directories using their URL. You can use the magic word `dotfiles` to allow or block the access to dot-files. The blocked files won't show in the admin interface. By default, `block dotfiles` is activated.
+ **allow_r** and **block_r** and variations of the previous options but you are able to use regular expressions with them. These regular expressions are used to match the URL, **not** the internal file path.


So, by **default** we have:

```
filemanager / {
    show           ./
    webdav         /webdav
    allow_new      true
    allow_edit     true
    allow_commands true
    allow_command  git
    allow_command  svn
    allow_command  hg
    block          dotfiles
}
```

As already mentioned, this extension should be used with [`basicauth`](https://caddyserver.com/docs/basicauth). If you do that, you will also be able to set permissions for different users using the following syntax:

```
filemanager {
    # You set the global configurations here and
    # all the users will inherit them.
    user1:
    # Here you can set specific settings for the 'user1'.
    # They will override the global ones for this specific user.
}
```

## Examples

Show the directory where Caddy is being executed at the root of the domain:

```
filemanager
```

Use only WebDAV:

```
filemanager {
    webdav /
}
```

Show the content of `foo` at the root of the domain:

```
filemanager {
    show foo/
}
```

Show the directory where Caddy is being executed at `/filemanager`:

```
filemanager /filemanager
```

Show the content of `foo` at `/bar`:

```
filemanager /bar{
    show   foo/
}
```

Now, a bit more complicated example. You have three users: an administrator, a manager and an editor. The administrator can do everything and has access to the commands `rm` and `mv` because he is a geeky. The manager, doesn't have access to commands, but can create and edit files. The editor can **only** edit files. He can't even create new ones, because he will only edit the files after the manager creates them for him. Both the editor and the manager won't have access to the financial folder. We would have:

```
basicauth /admin admin pass
basicauth /admin manager pass
basicauth /admin editor pass

filemanager /admin {
    show           ./
    allow_commands false
    admin:
    allow_commands true
    allow_command  rm
    allow_command  mv
    allow          dotfiles
    manager:
    block          /admin/financial
    editor:
    allow_new      false
    block          /admin/financial
}
```

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

## Developers

If you want to build Caddy from source with this plugin, you should take the following steps:

1. Download the Caddy source code (`go get github.com/mholt/caddy/caddy`)
2. Download the File Manager source code (`go get github.com/hacdias/caddy-filemanager`).
3. Navigate to the directory where File Manager's code is.
4. Run `go generate`. Otherwise, you will get an `undefined: Asset` error.
5. Navigate to the directory where Caddy's source code is and open the file `caddy/caddymain/run.go`.
6. Add the line `_ github.com/hacdias/caddy-filemanager` to the imports section.

Now you only need to build or install Caddy and you're good to go.

**Pre-commit Git Hook**

```
go generate
git add -A
```

### To edit CSS, JS and HTML

(In construction)
