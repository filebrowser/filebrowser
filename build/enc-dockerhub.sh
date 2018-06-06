#!/bin/sh

# - Run `$(command -v winpty) docker run --rm -it alpine sh -c "REPO='https://github.com/<user|organization>/<repo>'; USERNAME=<username>; PASSWORD=<password>; $(cat travis-enc.sh)"`
# - Check the log and copy the secure key to `.travis.yml`: https://docs.travis-ci.com/user/environment-variables/#Defining-encrypted-variables-in-.travis.yml

apk add -U --no-cache git openssh ruby ruby-dev libffi-dev build-base ruby-dev libc-dev libffi-dev linux-headers
gem install travis --no-rdoc --no-ri

git clone $REPO ./tmp-repo && cd tmp-repo

travis login --pro --auto
travis encrypt DOCKER_USER=$USERNAME --add --pro --debug
travis encrypt DOCKER_PASS=$PASSWORD --add --pro --debug

sh
