#!/bin/bash

# order of arguments: BBTOKEN, BBDOWNLOAD PATH, GIT COMMIT
# todo: make sure unzip is installed
# todo: handle sigterm gracefully, after this container should shut down

if [ $# -gt 0 ]; then
  args=("$@")
  bbtoken=${args[0]}
  gitclonepath=${args[1]}
  commit=${args[2]}
  git clone ${gitclonepath} /${commit}
  echo "cloned repo to /${commit}"
  cd /${commit}
  git checkout ${commit}
  echo "Finished with downloading source code"
else
    echo "no arguments were passed in"
    exit 1
fi




