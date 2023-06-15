## Gtstef fork of filebrowser

**Note: Intended to be used in docker only.**

This fork makes the following significant changes to filebrowser for origin:

 1. [x] Improves search to use index instead of filesystem.
    - [x] lightning fast
    - [ ] realtime results as you type
 1. [ ] Preview enhancements
    - preview default view is constrained to files subwindow,
    which can be toggled to fullscreen.
 1. [ ] Updated node version and dependencies
    - [ ] uses latest npm and node version
    - [ ] removes deprecated npm packages
 1. [ ] Improved routing
    - fixed bugs in original version
 1. [ ] Added authentication type
    - Using bearer token with remote authentication server

## About

filebrowser provides a file managing interface within a specified directory and it can be used to upload, delete, preview, rename and edit your files. It allows the creation of multiple users and each user can have its own directory. It can be used as a standalone app.

## Install

Using docker:

1. docker run:

```
docker run -it -v /path/to/folder:/srv -p 8080:80 gtstef/filebrowser:0.1.0
```

1. docker-compose:

  - with local storage

```
version: '3.7'
services:
  filebrowser:
    volumes:
      - '/path/to/folder:/srv'
      #- './database/:/database/'
      #- './config.json:/.filebrowser.json'
    ports:
      - '8080:80'
    image: gtstef/filebrowser:0.1.0
```

  - with network share

```
version: '3.7'
services:
  filebrowser:
    volumes:
      - 'nas:/srv'
      #- './database/:/database/'
      #- './config.json:/.filebrowser.json'
    ports:
      - '8080:80'
    image: gtstef/filebrowser:0.1.0
volumes:
  nas:
    driver_opts:
      type: cifs
      o: "username=myusername,password=mypassword,rw"
      device: "//fileshare/"
```

## Configuration

