#!/bin/sh

NAME=cwltail
ARCH=amd64

env GOOS=linux GOARCH=$ARCH go build -o $NAME-linux-$ARCH
env GOOS=darwin GOARCH=$ARCH go build -o $NAME-macos-$ARCH
