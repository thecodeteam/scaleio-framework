#! /usr/bin/env bash

rm -rf ./vendor
rm glide.lock

grep -R --exclude-dir vendor --exclude-dir .git --exclude-dir mesos --exclude build.sh TODO ./

glide up

GOOS=linux GOARCH=amd64 go build .
