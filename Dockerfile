FROM golang:1.15.6

RUN apt-get update \
  && apt-get install -y imagemagick \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /code
