# filebrowser cmds rm

Removes a command from an event hooker

## Synopsis

Removes a command from an event hooker. The provided index
is the same that's printed when you run 'cmds ls'. Note
that after each removal/addition, the index of the
commands change. So be careful when removing them after each
other.

You can also specify an optional parameter (index_end) so
you can remove all commands from 'index' to 'index_end',
including 'index_end'.

```
filebrowser cmds rm <event> <index> [index_end] [flags]
```

## Options

```
  -h, --help   help for rm
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
```

## See Also

* [filebrowser cmds](filebrowser-cmds.md)	 - Command runner management utility

