# Installation

File Browser is a single binary and can be used as a standalone executable. Although, some might prefer to use it with [Docker](https://www.docker.com) or [Caddy](https://caddyserver.com), which is a fantastic web server that enables HTTPS by default. Its installation is quite straightforward independently on which system you want to use.

## Quick Setup

The quickest way for beginners to start using File Browser is by opening your terminal and executing the following commands:

### Brew

```sh
brew tap filebrowser/tap
brew install filebrowser
filebrowser -r /path/to/your/files
```

### Unix

```sh
curl -fsSL https://raw.githubusercontent.com/filebrowser/get/master/get.sh | bash
filebrowser -r /path/to/your/files
```

### Windows

```sh
iwr -useb https://raw.githubusercontent.com/filebrowser/get/master/get.ps1 | iex
filebrowser -r /path/to/your/files
```

### Configuring

Done! It will bootstrap a database in which all the configurations and users are stored. Now, you can see on your command line the address in which your instance is running. You just need to go to that URL and use the following credentials:

* Username: `admin`
* Password: (printed in your console)

Although this is the fastest way to bootstrap an instance, we recommend you to take a look at other possible options, by checking `config init --help` and `config set --help`, to make the installation as safe and customized as it can be.

## Docker

File Browser is available as two different Docker images, which can be found on [Docker Hub](https://hub.docker.com/r/filebrowser/filebrowser).

### Alpine

```sh
docker run \
  -v /path/to/srv:/srv \
  -v /path/to/database:/database \
  -v /path/to/config:/config \
  -p 8080:80 \
  filebrowser/filebrowser
```

### s6 overlay

The `s6` image is based on LinuxServer and leverages the [s6-overlay](https://github.com/just-containers/s6-overlay) system for a standard, highly customizable image. It should be used as follows:

```shell
docker run \
  -v /path/to/srv:/srv \
  -v /path/to/database:/database \
  -v /path/to/config:/config \
  -e PUID=$(id -u) \
  -e PGID=$(id -g) \
  -p 8080:80 \
  filebrowser/filebrowser:s6
```

### Notes

Where:

- `/path/to/srv` contains the files root directory for File Browser
- `/path/to/config` contains a `settings.json` file
- `/path/to/database` contains a `filebrowser.db` file

Both `settings.json` and `filebrowser.db` will automatically be initialized if they don't exist.
