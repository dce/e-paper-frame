#!/bin/sh

docker run \
  --rm \
  -it \
  -v "$(pwd)":/code:delegated \
  -v frame-server-gopath:/go/src \
  -p 8080:80 \
  -w /code \
  golang:1.15.6 \
  /bin/bash
