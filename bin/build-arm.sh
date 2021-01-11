#!/bin/sh

env GOOS=linux GOARCH=arm GOARM=5 go build -o frame-server-arm
