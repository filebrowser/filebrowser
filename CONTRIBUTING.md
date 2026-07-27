# Building File Browser

This project is archived on 2026-09-01. Pull requests are not accepted and no further changes are merged. This document is kept as build documentation for anyone wishing to build the project from source.

## Project Structure

The backend side of the application is written in [Go](https://golang.org/), while the frontend (located on a subdirectory of the same name) is written in [Vue.js](https://vuejs.org/). Due to the tight coupling required by some features, basic knowledge of both Go and Vue.js is recommended.

* Learn Go: [https://github.com/golang/go/wiki/Learn](https://github.com/golang/go/wiki/Learn)
* Learn Vue.js: [https://vuejs.org/guide/introduction.html](https://vuejs.org/guide/introduction.html)

We encourage you to use git to manage your fork. To clone the main repository, just run:

```bash
git clone https://github.com/filebrowser/filebrowser
```

We use [Taskfile](https://taskfile.dev/) to manage the different processes (building, releasing, etc) automatically.

## Build

You can fully build the project in order to produce a binary by running:

```bash
task build
```

## Development

For development, there are a few things to have in mind.

### Frontend

We use [Node.js](https://nodejs.org/en/) on the frontend to manage the build process. Prepare the frontend environment:

```bash
# From the root of the repo, go to frontend/
cd frontend

# Install the dependencies
pnpm install
```

If you just want to develop the backend, you can create a static build of the frontend:

```bash
pnpm run build
```

If you want to develop the frontend, start a development server which watches for changes:

```bash
pnpm run dev
```

Please note that you need to access File Browser's interface through the development server of the frontend.

### Backend

First prepare the backend environment by downloading all required dependencies:

```bash
go mod download
```

You can now build or run File Browser as any other Go project:

```bash
# Build
go build

# Run
go run .
```

## Documentation

Documentation lives in [`docs`](docs) as plain Markdown and is no longer built into a site. The command line reference in [`docs/cli`](docs/cli) is generated from the commands themselves:

```bash
task docs:cli:generate
```

## Release

To make a release, just run:

```bash
task release
```

## Translations

The Transifex integration stopped on 2026-09-01 and translations submitted there no longer reach this repository. Locale files live in [`frontend/src/i18n`](frontend/src/i18n) and can be edited directly.

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

