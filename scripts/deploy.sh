#!/usr/bin/env bash

set -e  # exit on failure

## Only deploy if on correct branch.
printf "Branches: \$TRAVIS_BRANCH='%s'\n" $TRAVIS_BRANCH

## Login to Docker.
if ! echo "$DOCKER_PASS" | docker login -u "$DOCKER_USER" --password-stdin
  then exit 1
fi

make BRANCH="$TRAVIS_BRANCH" ci-deploy
