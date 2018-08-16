#! /bin/sh

set -e

cd $(dirname $0)

./docker_login.sh

for tag in `echo $(docker images filebrowser/filebrowser* | awk -F ' ' '{print $1 ":" $2}') | cut -d ' ' -f2-`; do
  if [ "$tag" = "REPOSITORY:TAG" ]; then break; fi
  docker push $tag
done

docker logout
