# filebrowser completion zsh

Generate the autocompletion script for zsh

## Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(filebrowser completion zsh)

To load completions for every new session, execute once:

### Linux:

	filebrowser completion zsh > "${fpath[1]}/_filebrowser"

### macOS:

	filebrowser completion zsh > $(brew --prefix)/share/zsh/site-functions/_filebrowser

You will need to start a new shell for this setup to take effect.


```
filebrowser completion zsh [flags]
```

## Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
```

## See Also

* [filebrowser completion](filebrowser-completion.md)	 - Generate the autocompletion script for the specified shell

