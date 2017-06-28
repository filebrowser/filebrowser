#!/bin/sh

# Script to open a browser to current branch
# Repo formats:
# ssh   git@github.com:richo/gh_pr.git
# http  https://richoH@github.com/richo/gh_pr.git
# git   git://github.com/richo/gh_pr.git

username=`git config --get github.user`

get_repo() {
    git remote -v | grep ${@:-$username} | while read remote; do
      if repo=`echo $remote | grep -E -o "git@github.com:[^ ]*"`; then
          echo $repo | sed -e "s/^git@github\.com://" -e "s/\.git$//"
          exit 1
      fi
      if repo=`echo $remote | grep -E -o "https?://([^@]*@)?github.com/[^ ]*\.git"`; then
          echo $repo | sed -e "s|^https?://||" -e "s/^.*github\.com\///" -e "s/\.git$//"
          exit 1
      fi
      if repo=`echo $remote | grep -E -o "git://github.com/[^ ]*\.git"`; then
          echo $repo | sed -e "s|^git://github.com/||" -e "s/\.git$//"
          exit 1
      fi
    done

    if [ $? -eq 0 ]; then
        echo "Couldn't find a valid remote" >&2
        exit 1
    fi
}

echo ${#x[@]}

if repo=`get_repo $@`; then
    branch=`git symbolic-ref HEAD 2>/dev/null`
    echo "http://github.com/$repo/pull/new/${branch##refs/heads/}"
else
    exit 1
fi
