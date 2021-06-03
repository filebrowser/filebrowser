#!/bin/sh

cd /
chmod 777 filebrowser
./filebrowser config init
./filebrowser config set --auth.method=noauth
./filebrowser users add admin ${PASSWORD}
./filebrowser -a 0.0.0.0 -p 80 -b /srv
