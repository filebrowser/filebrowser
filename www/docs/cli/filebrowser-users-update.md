# filebrowser users update

Updates an existing user

## Synopsis

Updates an existing user. Set the flags for the
options you want to change.

```
filebrowser users update <id|username> [flags]
```

## Options

```
      --aceEditorTheme string   ace editor's syntax highlighting theme for users
      --commands strings        a list of the commands a user can execute
      --dateFormat              use date format (true for absolute time, false for relative)
  -h, --help                    help for update
      --hideDotfiles            hide dotfiles
      --locale string           locale for users (default "en")
      --lockPassword            lock password
  -p, --password string         new password
      --perm.admin              admin perm for users
      --perm.create             create perm for users (default true)
      --perm.delete             delete perm for users (default true)
      --perm.download           download perm for users (default true)
      --perm.execute            execute perm for users (default true)
      --perm.modify             modify perm for users (default true)
      --perm.rename             rename perm for users (default true)
      --perm.share              share perm for users (default true)
      --redirectAfterCopyMove   redirect to destination after copy/move
      --scope string            scope for users (default ".")
      --singleClick             use single clicks only
      --sorting.asc             sorting by ascending order
      --sorting.by string       sorting mode (name, size or modified) (default "name")
  -u, --username string         new username
      --viewMode string         view mode for users (default "list")
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
```

## See Also

* [filebrowser users](filebrowser-users.md)	 - Users management utility

