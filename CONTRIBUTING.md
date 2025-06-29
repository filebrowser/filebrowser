# Contributing

If you're interested in contributing to this project, this is the best place to start. Before contributing to this project, please take a bit of time to read our [Code of Conduct](code-of-conduct.md). Also, note that this project is open-source and licensed under [Apache License 2.0](LICENSE).

## Project Structure

The backend side of the application is written in [Go](https://golang.org/), while the frontend (located on a subdirectory of the same name) is written in [Vue.js](https://vuejs.org/). Due to the tight coupling required by some features, basic knowledge of both Go and Vue.js is recommended.

* Learn Go: [https://github.com/golang/go/wiki/Learn](https://github.com/golang/go/wiki/Learn)
* Learn Vue.js: [https://vuejs.org/guide/introduction.html](https://vuejs.org/guide/introduction.html)

We encourage you to use git to manage your fork. To clone the main repository, just run:

```bash
git clone https://github.com/filebrowser/filebrowser
```

## Build

### Frontend

We are using [Node.js](https://nodejs.org/en/) on the frontend to manage the build process. The steps to build it are:

```bash
# From the root of the repo, go to frontend/
cd frontend

# Install the dependencies
pnpm install

# Build the frontend
pnpm run build
```

This will install the dependencies and build the frontend so you can then embed it into the Go app. Although, if you want to play with it, you'll get bored of building it after every change you do. So, you can run the command below to watch for changes:

```bash
pnpm run dev
```

### Backend

First of all, you need to download the required dependencies. We are using the built-in `go mod` tool for dependency management. To get the modules, run:

```bash
go mod download
```

The magic of File Browser is that the static assets are bundled into the final binary. For that, we use [Go embed.FS](https://golang.org/pkg/embed/). The files from `frontend/dist` will be embedded during the build process.

To build File Browser is just like any other Go program:

```bash
go build
```

To create a development build use the "dev" tag, this way the content inside the frontend folder will not be embedded in the binary but will be reloaded at every change:

```bash
go build -tags dev
```

## Translations

Translations are managed on Transifex, which is an online website where everyone can contribute and translate strings for our project. It automatically syncs with the main language file \(in English\) and,, for you to contribute, you just need to:

1. Go to our Transifex web page: [app.transifex.com/file-browser/file-browser](https://app.transifex.com/file-browser/file-browser/)
2. Click on **Join the project** and pick your language. We'll accept you as soon as possible. If you're language is not on the list, please request it via the interface.

Translations are automatically pushed to GitHub via an integration.

## Authentication Provider

To build a new authentication provider, you need to implement the [Auther interface](https://github.com/filebrowser/filebrowser/blob/master/auth/auth.go), whose method will be called on the login page after the user has submitted their login data.

```go
// Auther is the authentication interface.
type Auther interface {
    // Auth is called to authenticate a request.
    Auth(r *http.Request, s *users.Storage, root string) (*users.User, error)
}
```

After implementing the interface you should:

1. Add it to [`auth` directory](https://github.com/filebrowser/filebrowser/blob/master/auth).
2. Add it to the [configuration parser](https://github.com/filebrowser/filebrowser/blob/master/cmd/config.go) for the CLI.
3. Add it to the [`authBackend.Get`](https://github.com/filebrowser/filebrowser/blob/master/storage/bolt/auth.go).

If you need to add more flags, please update the function `addConfigFlags`.

