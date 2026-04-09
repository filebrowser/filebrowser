# filebrowser config export

Export the configuration to a file

## Synopsis

Export the configuration to a file. The path must be for a
json or yaml file. This exported configuration can be changed,
and imported again with 'config import' command.

```
filebrowser config export <path> [flags]
```

## Options

```
  -h, --help   help for export
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
```

## See Also

* [filebrowser config](filebrowser-config.md)	 - Configuration management utility

