#!/bin/bash
# deploy my cross-arm go binary to a raspberry pi to run a server

set -e
set -x

REMOTE=pi@rpi.local

ssh $REMOTE "killall go-weather" | true
ssh $REMOTE "mkdir -p ~/ad/go-weather"

scp go-weather $REMOTE:ad/go-weather/
scp apiConfig $REMOTE:ad/go-weather/

ssh $REMOTE "cd ~/ad/go-weather/ && ./go-weather"
