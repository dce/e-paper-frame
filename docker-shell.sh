#!/bin/sh

docker run \
  --rm \
  -it \
  --mount type=bind,source="$(pwd)",target=/code \
  -p 8080:8080 \
  golang:1.15.6 \
  /bin/bash
