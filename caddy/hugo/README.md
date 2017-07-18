# hugo - a caddy plugin

[![community](https://img.shields.io/badge/community-forum-ff69b4.svg?style=flat-square)](https://caddy.community)

hugo fills the gap between Hugo and the browser. [Hugo][6] is an easy and fast static website generator. This plugin fills the gap between Hugo and the end-user, providing you a web interface to manage the whole website.

Using this plugin, you won't need to have your own computer to edit posts, neither regenerate your static website, because you can do all of that just through your browser. It is an implementation of [hacdias/filemanager][1] library.

**Requirements:** you need to have the hugo executable in your PATH. You can download it from its [official page][6].

## Get Started

To start using this plugin you just need to go to the [download Caddy page][3] and choose `http.hugo` in the directives section. For further information on how Caddy works refer to [its documentation][4].

The default credentials are `admin` for both the user and the password. It is highy recommended to change them after logging in for the first time and to use HTTPS. You can create more users and define their own permissions using the web interface.

## Syntax

```
hugo [directory] [admin] {
    database path
}
```

+ `directory` is the path, relative or absolute to the directory of your Hugo files. Defaults to `./`.
+ `admin` is the URL path where you will access the admin interface. Defaults to `/admin`.
+ `path` is the database path where the settings will be stored. By default, the settings will be stored on [`.caddy`][5] folder.

## Database

By default the database will be stored on [`.caddy`][5] directory, in a sub-directory called `hugo`. Each file name is an hash of the combination of the host and the base URL.

If you don't set a database path, you will receive a warning like this:

> [WARNING] A database is going to be created for your File Manager instace at ~/.caddy/hugo/xxx.db. It is highly recommended that you set the 'database' option to 'xxx.db'

Why? If you don't set a database path and you change the host or the base URL, your settings will be reseted. So it is *highly* recommended to set this option.

When you set a relative path, such as `xxxxxxxxxx.db`, it will always be relative to `.caddy/hugo` directory. Although, you may also use an absolute path if you wish to store the database in other place.

## Examples

Manage the current working directory's Hugo website at `/admin`.

```
hugo {
  database myinstance.db
}
```

Manage the Hugo website located at `/var/www/mysite` at `/admin`.

```
hugo /var/www/mysite {
  database myinstance.db
}
```

Manage the Hugo website located at `/var/www/mysite` at `/private`.

```
hugo /var/www/mysite /private {
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
[6]:http://gohugo.io
