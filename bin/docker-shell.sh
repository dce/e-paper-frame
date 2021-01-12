#!/bin/sh

docker run \
  --rm \
  -it \
  -v "$(pwd)":/code:delegated \
  -v frame-server-gopath:/go/src \
  -p 8080:80 \
  frame-server \
  /bin/bash
