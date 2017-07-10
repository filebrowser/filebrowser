# filemanager - a caddy plugin

filemanager provides WebDAV features and a file managing interface within the specified directory and it can be used to upload, delete, preview, rename and edit your files within that directory. It is an implementation of [hacdias/filemanager](https://github.com/hacdias/filemanager) library.

Note that if you are handling large files you might run into troubles due to the defaults of [`timeouts`](https://caddyserver.com/docs/timeouts) plugin. Check its [documentation](https://caddyserver.com/docs/timeouts) to learn more about that plugin. For information about the working of filemanager itself, go to the [main repository](https://github.com/hacdias/filemanager).

## Get Started

To start using this plugin you just need to go to the [download Caddy page](https://caddyserver.com/download) and choose `filemanager` in the directives section. For further information on how Caddy works refer to [its documentation](https://caddyserver.com/docs).

The default credentials are `admin` for both the user and the password. It is highy recommended to change them after logging in for the first time and to use HTTPS. You can create more users and define their own permissions using the web interface.

## Syntax

```
filemanager [baseurl] [scope] {
    database  path
}
```

`baseurl` is the URL path where you will access File Manager. Defaults to `/`.

`scope` is the path, relative or absolute, to the directory you want to browse in. This value will be used for the creation of the first user. Defaults to `./`.

`path` is the database path where the settings will be stored. By default, the database will be stored on `.caddy/filemanager` folder and its name will be an hashed combination of the host and the `baseurl`. If you use a relative path it will be relative to `.caddy/filemanager`. Despite being optional, it is **highly** recommended to set this option in order to keep the settings when you change the `baseurl` and/or the hostname.

## Examples

Show the directory where Caddy is being executed at the root of the domain:

```
filemanager
```


Show the content of `foo` at the root of the domain:

```
filemanager / ./foo
```

Show the directory where Caddy is being executed at `/filemanager`:

```
filemanager /filemanager
```

Show the content of `foo` at `/bar`:

```
filemanager /bar /show
```
