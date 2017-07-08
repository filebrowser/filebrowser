# filemanager - a caddy plugin

filemanager provides WebDAV features and a file managing interface within the specified directory and it can be used to upload, delete, preview, rename and edit your files within that directory. It is an implementation of [hacdias/filemanager](https://github.com/hacdias/filemanager) library.

Note that if you are handling large files you might run into troubles due to the defaults of [`timeouts`](https://caddyserver.com/docs/timeouts) plugin. Check its [documentation](https://caddyserver.com/docs/timeouts) to learn more about that plugin.

For information about the working of filemanager itself, go to the [main repository](https://github.com/hacdias/filemanager).

## Get Started

To start using this plugin you just need to go to the [download Caddy page](https://caddyserver.com/download) and choose `filemanager` in the directives section. For further information on how Caddy works refer to [its documentation](https://caddyserver.com/docs).

## Syntax

```
filemanager [baseurl] [scope] {
    database  path
}
```

All of the options above are optional.

+ **baseurl** is the URL where you will access the File Manager interface. Defaults to `/`.
+ **scope** is the path, relative or absolute, to the directory you want to browse in. Defaults to `./`.
+ **path** is the database path where File Manager will store the settings that aren't included in the Caddyfile. By default, the database will be stored on `.caddy` folder and its name will be an hashed combination of the host and the `baseurl`. It is **highly** recommended to set this option. Otherwise, whenever you change the host or the baseurl, your settings will be lost or you will need to point to the previous database.

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
}
```
