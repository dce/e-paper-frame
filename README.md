# E-Paper Frame

This is the source code for an e-paper/Raspberry Pi frame I made.

The main source is `frame-server.go`. You'll need to `go get github.com/dce/rpi/epd7in5` to build the project, but otherwise it should be pretty straightforward. I've included a couple scripts in `bin` to making development easier:

* `docker-shell.sh` spins up a session you can compile the project in
* `build.sh` compiles the project
* `build-arm.sh` compiles the project for running on the Pi
* `deploy.sh` logs into the Pi, stops the program, uploads the new version, and starts the program

In `etc`, you'll find a couple useful utilities:

* `frame-server.service` is a `systemd` configuration to control the program
* `random-photo` is a script you can put in `/etc/cron-hourly` that fetches new photos from S3 and displays a random photo on the Pi
