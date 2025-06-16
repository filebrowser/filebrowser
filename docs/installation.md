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

File Browser is also available as a Docker image. You can find it on [Docker Hub](https://hub.docker.com/r/filebrowser/filebrowser). The usage is as follows:

### Alpine

```sh
docker run \
  -v /path/to/root:/srv \
  -v /path/filebrowser.db:/database.db \
  -v /path/.filebrowser.json:/.filebrowser.json \
  -u $(id -u):$(id -g) \
  -p 8080:80 \
  filebrowser/filebrowser
```

### LinuxServer

```shell
docker run \
  -v /path/to/root:/srv \
  -v /path/to/filebrowser.db:/database/filebrowser.db \
  -v /path/to/settings.json:/config/settings.json \
  -e PUID=$(id -u) \
  -e PGID=$(id -g) \
  -p 8080:80 \
  filebrowser/filebrowser:s6
```

By default, we already have a [configuration file with some defaults](../docker/root/defaults/settings.json) so you can just mount the root and the database. Although you can overwrite by mounting a directory with a new config file. If you don't already have a database file, make sure to create a new empty file under the path you specified. Otherwise, Docker will create an empty folder instead of an empty file, resulting in an error when mounting the database into the container.
