# filemanager - a caddy plugin

filemanager provides a file managing interface within a specified directory and it can be used to upload, delete, preview, rename and edit your files. It allows the creation of multiple users and each user can have its own directory.bIt is an implementation of [hacdias/filemanager][1] library.

## Get Started

To start using this plugin you just need to go to the [download Caddy page][3] and choose `http.filemanager` in the directives section. For further information on how Caddy works refer to [its documentation][4].

The default credentials are `admin` for both the user and the password. It is highy recommended to change them after logging in for the first time and to use HTTPS. You can create more users and define their own permissions using the web interface.

For information about the working of filemanager itself, go to the [main repository](https://github.com/hacdias/filemanager).

## Syntax

```
filemanager [baseurl] [scope] {
    database  path
}
```

+ `baseurl` is the URL path where you will access File Manager. Defaults to `/`.
+ `scope` is the path, relative or absolute, to the directory you want to browse in. This value will be used for the creation of the first user. Defaults to `./`.
+ `path` is the database path where the settings will be stored. By default, the settings will be stored on [`.caddy`][5] folder.

## Database

By default the database will be stored on [`.caddy`][5] directory, in a sub-directory called `filemanager`. Each file name is an hash of the combination of the host and the base URL.

If you don't set a database path, you will receive a warning like this:

> [WARNING] A database is going to be created for your File Manager instace at ~/.caddy/filemanager/xxx.db. It is highly recommended that you set the 'database' option to 'xxx.db'

Why? If you don't set a database path and you change the host or the base URL, your settings will be reseted. So it is *highly* recommended to set this option.

When you set a relative path, such as `xxxxxxxxxx.db`, it will always be relative to `.caddy/filemanager` directory. Although, you may also use an absolute path if you wish to store the database in other place.

## Examples

Show the directory where Caddy is being executed at the root of the domain:

```
filemanager {
  database myinstance.db
}
```


Show the content of `foo` at the root of the domain:

```
filemanager / ./foo {
  database myinstance.db
}
```

Show the directory where Caddy is being executed at `/filemanager`:

```
filemanager /filemanager {
  database myinstance.db
}
```

Show the content of `foo` at `/bar`:

```
filemanager /bar /show {
  database myinstance.db
}
```

## Known Issues

If you are having troubles **handling large files** you might need to check out the [`timeouts`][2] plugin, which can be used to change the default HTTP Timeouts.

[1]:https://github.com/hacdias/filemanager
[2]:https://caddyserver.com/docs/timeouts
[3]:https://caddyserver.com/download
[4]:https://caddyserver.com/docs
[5]:https://caddyserver.com/docs/automatic-https#dot-caddy
