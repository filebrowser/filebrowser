# filebrowser

A stylish web-based file browser

## Synopsis

File Browser CLI lets you create the database to use with File Browser,
manage your users and all the configurations without accessing the
web interface.

If you've never run File Browser, you'll need to have a database for
it. Don't worry: you don't need to setup a separate database server.
We're using Bolt DB which is a single file database and all managed
by ourselves.

For this specific command, all the flags you have available (except
"config" for the configuration file), can be given either through
environment variables or configuration files.

If you don't set "config", it will look for a configuration file called
.filebrowser.{json, toml, yaml, yml} in the following directories:

- ./
- $HOME/
- /etc/filebrowser/

The precedence of the configuration values are as follows:

- flags
- environment variables
- configuration file
- database values
- defaults

The environment variables are prefixed by "FB_" followed by the option
name in caps. So to set "database" via an env variable, you should
set FB_DATABASE.

Also, if the database path doesn't exist, File Browser will enter into
the quick setup mode and a new database will be bootstrapped and a new
user created with the credentials from options "username" and "password".

```
filebrowser [flags]
```

## Options

```
  -a, --address string                     address to listen on (default "127.0.0.1")
  -b, --baseurl string                     base url
      --cache-dir string                   file cache directory (disabled if empty)
  -t, --cert string                        tls certificate
  -c, --config string                      config file path
  -d, --database string                    database path (default "./filebrowser.db")
      --disable-exec                       disables Command Runner feature (default true)
      --disable-preview-resize             disable resize of image previews
      --disable-thumbnails                 disable image thumbnails
      --disable-type-detection-by-header   disables type detection by reading file headers
  -h, --help                               help for filebrowser
      --img-processors int                 image processors count (default 4)
  -k, --key string                         tls key
  -l, --log string                         log output (default "stdout")
      --noauth                             use the noauth auther when using quick setup
      --password string                    hashed password for the first user when using quick config
  -p, --port string                        port to listen on (default "8080")
  -r, --root string                        root to prepend to relative paths (default ".")
      --socket string                      socket to listen to (cannot be used with address, port, cert nor key flags)
      --socket-perm uint32                 unix socket file permissions (default 438)
      --token-expiration-time string       user session timeout (default "2h")
      --username string                    username for the first user when using quick config (default "admin")
```

## See Also

* [filebrowser cmds](filebrowser-cmds.md)	 - Command runner management utility
* [filebrowser completion](filebrowser-completion.md)	 - Generate the autocompletion script for the specified shell
* [filebrowser config](filebrowser-config.md)	 - Configuration management utility
* [filebrowser hash](filebrowser-hash.md)	 - Hashes a password
* [filebrowser rules](filebrowser-rules.md)	 - Rules management utility
* [filebrowser users](filebrowser-users.md)	 - Users management utility
* [filebrowser version](filebrowser-version.md)	 - Print the version number

