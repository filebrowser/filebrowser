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

For this command, all flags are available as environmental variables,
except for "--config", which specifies the configuration file to use.
The environment variables are prefixed by "FB_" followed by the flag name in
UPPER_SNAKE_CASE. For example, the flag "--disablePreviewResize" is available
as FB_DISABLE_PREVIEW_RESIZE.

If "--config" is not specified, File Browser will look for a configuration
file named .filebrowser.{json, toml, yaml, yml} in the following directories:

- ./
- $HOME/
- /etc/filebrowser/

**Note:** Only the options listed below can be set via the config file or
environment variables. Other configuration options live exclusively in the
database and so they must be set by the "config set" or "config
import" commands.

The precedence of the configuration values are as follows:

- Flags
- Environment variables
- Configuration file
- Database values
- Defaults

Also, if the database path doesn't exist, File Browser will enter into
the quick setup mode and a new database will be bootstrapped and a new
user created with the credentials from options "username" and "password".

```
filebrowser [flags]
```

## Options

```
  -a, --address string                 address to listen on (default "127.0.0.1")
  -b, --baseURL string                 base url
      --cacheDir string                file cache directory (disabled if empty)
  -t, --cert string                    tls certificate
  -c, --config string                  config file path
  -d, --database string                database path (default "./filebrowser.db")
      --disableExec                    disables Command Runner feature (default true)
      --disableImageResolutionCalc     disables image resolution calculation by reading image files
      --disablePreviewResize           disable resize of image previews
      --disableThumbnails              disable image thumbnails
      --disableTypeDetectionByHeader   disables type detection by reading file headers
  -h, --help                           help for filebrowser
      --imageProcessors int            image processors count (default 4)
  -k, --key string                     tls key
  -l, --log string                     log output (default "stdout")
      --noauth                         use the noauth auther when using quick setup
      --password string                hashed password for the first user when using quick setup
  -p, --port string                    port to listen on (default "8080")
  -r, --root string                    root to prepend to relative paths (default ".")
      --socket string                  socket to listen to (cannot be used with address, port, cert nor key flags)
      --socketPerm uint32              unix socket file permissions (default 438)
      --tokenExpirationTime string     user session timeout (default "2h")
      --username string                username for the first user when using quick setup (default "admin")
```

## See Also

* [filebrowser cmds](filebrowser-cmds.md)	 - Command runner management utility
* [filebrowser completion](filebrowser-completion.md)	 - Generate the autocompletion script for the specified shell
* [filebrowser config](filebrowser-config.md)	 - Configuration management utility
* [filebrowser hash](filebrowser-hash.md)	 - Hashes a password
* [filebrowser rules](filebrowser-rules.md)	 - Rules management utility
* [filebrowser users](filebrowser-users.md)	 - Users management utility
* [filebrowser version](filebrowser-version.md)	 - Print the version number

