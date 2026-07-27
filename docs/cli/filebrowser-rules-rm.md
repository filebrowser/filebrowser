# filebrowser rules rm

Remove a global rule or user rule

## Synopsis

Remove a global rule or user rule. The provided index
is the same that's printed when you run 'rules ls'. Note
that after each removal/addition, the index of the
commands change. So be careful when removing them after each
other.

You can also specify an optional parameter (index_end) so
you can remove all commands from 'index' to 'index_end',
including 'index_end'.

```
filebrowser rules rm <index> [index_end] [flags]
```

## Options

```
  -h, --help         help for rm
      --index uint   index of rule to remove
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
  -i, --id uint           id of user to which the rules apply
  -u, --username string   username of user to which the rules apply
```

## See Also

* [filebrowser rules](filebrowser-rules.md)	 - Rules management utility

