#!/bin/sh

if [ -z ${IP+x} ]; then
  IP=192.168.4.61
  echo "Using default IP address of $IP"
fi

ssh pi@$IP "sudo systemctl stop frame-server" \
  && scp frame-server-arm pi@$IP:/home/pi/frame \
  && ssh pi@$IP "sudo systemctl start frame-server"
