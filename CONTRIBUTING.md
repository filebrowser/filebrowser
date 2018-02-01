# Contributing

If you want to contribute or want to build the code from source, you will need to have the most recent version of Go and, if you want to change the static assets (JS, CSS, ...), Node.js installed on your computer. To start developing, you just need to do the following:

1. `go get github.com/hacdias/filemanager/cmd/filemanager`
2. `cd $GOPATH/src/github.com/hacdias/filemanager`
3. `npm install`
4. `npm run dev` - regenerates the static assets automatically
5. `go install github.com/hacdias/filemanager/cmd/filemanager`
6. Execute `$GOPATH/bin/filemanager`

The steps 3 and 4 are only required **if you want to develop the front-end**. Otherwise, you can ignore them. Before pulling, if you made any change on assets folder, you must run the `build.sh` script on the root of this repository.

If you are using this as a Caddy plugin, you should use its [official instructions for plugins](https://github.com/mholt/caddy/wiki/Extending-Caddy#2-plug-in-your-plugin) and import `github.com/hacdias/filemanager/caddy/filemanager`.
