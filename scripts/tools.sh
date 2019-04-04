#!/usr/bin/env bash

## Options:
DOCKER_COMPOSE_VERSION=1.23.2


set -e  # exit on failure

## Configure $BINPATH for third-party binaries.
mkdir -p $BINPATH
echo "Contents of $BINPATH:" && ls -l $BINPATH

## Install docker-compose.
if [ ! -x ${BINPATH}/docker-compose ]; then
  echo "Installing docker-compose..."
  VERSION="docker-compose-$(uname -s)-$(uname -m)"
  curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/${VERSION}" > docker-compose
  chmod +x docker-compose
  mv docker-compose $BINPATH
  echo done
fi
echo "docker-compose: $(docker-compose version)"
