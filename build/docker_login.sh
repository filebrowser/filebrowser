#!/bin/sh

set -e

# init key for pass
gpg --batch --gen-key <<-EOF
%echo Generating a standard key
Key-Type: DSA
Key-Length: 1024
Subkey-Type: ELG-E
Subkey-Length: 1024
Name-Real: Meshuggah Rocks
Name-Email: meshuggah@example.com
Expire-Date: 0
# Do a commit here, so that we can later print "done" :-)
%commit
%echo done
EOF

key=$(gpg --no-auto-check-trustdb --list-secret-keys | grep ^sec | cut -d/ -f2 | cut -d" " -f1)
pass init $key

if [ "$(command -v docker-credential-pass)" = "" ]; then
  docker run --rm -itv /usr/local/bin:/src filebrowser/dev sh -c "cp /go/bin/docker-credential-pass /src"
fi

echo "$DOCKER_PASS" | docker login -u "$DOCKER_USER" --password-stdin
