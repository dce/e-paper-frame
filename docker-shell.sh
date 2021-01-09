#!/bin/sh

docker run \
  --rm \
  -it \
  -v "$(pwd)":/code:delegated \
  -v frame-server-gopath:/go/src \
  -w /code \
  -p 8080:80 \
  golang:1.15.6 \
  /bin/bash
