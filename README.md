> [!WARNING]
> **File Browser is being archived on 2026-10-01.** On that date this repository becomes read-only. There will be no further releases, and no security fixes, including for the [known unfixed issues](#security) listed below.
>
> Existing releases and Docker images stay online and will not be withdrawn. All documentation lives in this repository under [`www/docs`](www/docs).

<p align="center">
  <img src="https://raw.githubusercontent.com/filebrowser/filebrowser/master/branding/banner.png" width="550"/>
</p>

[![Build](https://github.com/filebrowser/filebrowser/actions/workflows/ci.yaml/badge.svg)](https://github.com/filebrowser/filebrowser/actions/workflows/ci.yaml)
[![Version](https://img.shields.io/github/release/filebrowser/filebrowser.svg)](https://github.com/filebrowser/filebrowser/releases/latest)

File Browser provides a file managing interface within a specified directory and it can be used to upload, delete, preview and edit your files. It is a **create-your-own-cloud**-kind of software where you can just install it on your server, direct it to a path and access your files through a nice web interface.

## Documentation

Documentation on how to install, configure, and build this project lives in [`www/docs`](www/docs) in this repository.

## Project Status

This project is a finished product which fulfills its goal: be a single binary web File Browser which can be run by anyone anywhere. It has been in **maintenance-only** mode, and is now being wound down.

- The repository is archived on **2026-10-01** and becomes read-only.
- One last planned release ships before that date. It contains no functional changes, only a wind-down notice.
- After that date there are no further releases, bug fixes, or security fixes.
- Existing releases and Docker images stay online and are not withdrawn.

Please read [@hacdias' personal reflection](https://hacdias.com/2026/03/11/filebrowser/) on the project status.

## Security

No further security fixes will be released. Two known issue classes remain unaddressed and will not be fixed:

- **Command execution, runner, and hooks** (opt-in, disabled by default, [#5199](https://github.com/filebrowser/filebrowser/issues/5199))
- **Session and JWT handling** ([#5216](https://github.com/filebrowser/filebrowser/issues/5216))

Previously published advisories are listed under [security advisories](https://github.com/filebrowser/filebrowser/security/advisories). Reporting instructions, which apply until 2026-10-01, are in [SECURITY.md](SECURITY.md).

If you keep running File Browser after the archive date, treat it as unmaintained software:

- **Do not expose it directly to the internet.** Put it behind a reverse proxy that terminates TLS and performs its own authentication.
- **Keep the command runner disabled.** It is off by default, so leave it off. See [#5199](https://github.com/filebrowser/filebrowser/issues/5199) and [`www/docs/command-execution.md`](www/docs/command-execution.md).
- **Run it unprivileged, inside a container**, with only the directory you intend to serve mounted into it.

## Contributing

New features are not being accepted. Bug reports and bug-fix pull requests are still looked at until 2026-10-01, but there is no guarantee that they will be merged.

[CONTRIBUTING.md](CONTRIBUTING.md) documents how to build and develop the project, which remains useful to anyone forking it.

## License

[Apache License 2.0](LICENSE) © File Browser Contributors
