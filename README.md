> [!WARNING]
> 
> **File Browser is archived on 2026-09-01**. The last planned release has already shipped. There will be no further releases, bug fixes, or security fixes.   

<p align="center">
  <img src="./branding/banner.png" width="550"/>
</p>

File Browser provides a file managing interface within a specified directory and it can be used to upload, delete, preview and edit your files. It is a **create-your-own-cloud**-kind of software where you can just install it on your server, direct it to a path and access your files through a nice web interface.

**Background:** [Update On File Browser](https://hacdias.com/2026/03/11/filebrowser/), March 2026.

## Security

Published advisories are listed under [security advisories](https://github.com/filebrowser/filebrowser/security/advisories), and reporting instructions are in [SECURITY.md](SECURITY.md). This project has known security issue classes that remain unaddressed and will not be fixed:

- **Command execution, runner, and hooks** (opt-in, disabled by default, [#5199](https://github.com/filebrowser/filebrowser/issues/5199))
- **Session and JWT handling** ([#5216](https://github.com/filebrowser/filebrowser/issues/5216))

If you keep running File Browser, treat it as unmaintained software:

- **Do not expose it directly to the internet.** Put it behind a reverse proxy that terminates TLS and performs its own authentication.
- **Keep the command runner disabled.** It is off by default, so leave it off. See [#5199](https://github.com/filebrowser/filebrowser/issues/5199) and [`docs/command-execution.md`](docs/command-execution.md).
- **Run it unprivileged, inside a container**, with only the directory you intend to serve mounted into it.

## Documentation

Documentation on how to install, configure, and build this project lives in [`docs`](docs) in this repository.

[CONTRIBUTING.md](CONTRIBUTING.md) documents how to build and develop the project, which remains useful to anyone forking it.

## License

[Apache License 2.0](LICENSE) © File Browser Contributors
