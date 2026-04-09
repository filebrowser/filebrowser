# filebrowser config init

Initialize a new database

## Synopsis

Initialize a new database to use with File Browser. All of
this options can be changed in the future with the command
'filebrowser config set'. The user related flags apply
to the defaults when creating new users and you don't
override the options.

```
filebrowser config init [flags]
```

## Options

```
      --aceEditorTheme string            ace editor's syntax highlighting theme for users
  -a, --address string                   address to listen on (default "127.0.0.1")
      --auth.command string              command for auth.method=hook
      --auth.header string               HTTP header for auth.method=proxy
      --auth.logoutPage string           url of custom logout page
      --auth.method string               authentication type (default "json")
  -b, --baseURL string                   base url
      --branding.color string            set the theme color
      --branding.disableExternal         disable external links such as GitHub links
      --branding.disableUsedPercentage   disable used disk percentage graph
      --branding.files string            path to directory with images and custom styles
      --branding.name string             replace 'File Browser' by this name
      --branding.theme string            set the theme
  -t, --cert string                      tls certificate
      --commands strings                 a list of the commands a user can execute
      --createUserDir                    generate user's home directory automatically
      --dateFormat                       use date format (true for absolute time, false for relative)
      --dirMode string                   mode bits that new directories are created with (default "0o750")
      --disableExec                      disables Command Runner feature (default true)
      --disableImageResolutionCalc       disables image resolution calculation by reading image files
      --disablePreviewResize             disable resize of image previews
      --disableThumbnails                disable image thumbnails
      --disableTypeDetectionByHeader     disables type detection by reading file headers
      --fileMode string                  mode bits that new files are created with (default "0o640")
  -h, --help                             help for init
      --hideDotfiles                     hide dotfiles
      --hideLoginButton                  hide login button from public pages
  -k, --key string                       tls key
      --locale string                    locale for users (default "en")
      --lockPassword                     lock password
  -l, --log string                       log output (default "stdout")
      --minimumPasswordLength uint       minimum password length for new users (default 12)
      --perm.admin                       admin perm for users
      --perm.create                      create perm for users (default true)
      --perm.delete                      delete perm for users (default true)
      --perm.download                    download perm for users (default true)
      --perm.execute                     execute perm for users (default true)
      --perm.modify                      modify perm for users (default true)
      --perm.rename                      rename perm for users (default true)
      --perm.share                       share perm for users (default true)
  -p, --port string                      port to listen on (default "8080")
      --recaptcha.host string            use another host for ReCAPTCHA. recaptcha.net might be useful in China (default "https://www.google.com")
      --recaptcha.key string             ReCaptcha site key
      --recaptcha.secret string          ReCaptcha secret
      --redirectAfterCopyMove            redirect to destination after copy/move
  -r, --root string                      root to prepend to relative paths (default ".")
      --scope string                     scope for users (default ".")
      --shell string                     shell command to which other commands should be appended
  -s, --signup                           allow users to signup
      --singleClick                      use single clicks only
      --socket string                    socket to listen to (cannot be used with address, port, cert nor key flags)
      --sorting.asc                      sorting by ascending order
      --sorting.by string                sorting mode (name, size or modified) (default "name")
      --tokenExpirationTime string       user session timeout (default "2h")
      --tus.chunkSize uint               the tus chunk size (default 10485760)
      --tus.retryCount uint16            the tus retry count (default 5)
      --viewMode string                  view mode for users (default "list")
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
```

## See Also

* [filebrowser config](filebrowser-config.md)	 - Configuration management utility

