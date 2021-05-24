#!/bin/bash

chmod 777 filebrowser
./filebrowser config init
./filebrowser users add admin ${PASSWORD}
./filebrowser -a 0.0.0.0 -p 80 -b /srv
