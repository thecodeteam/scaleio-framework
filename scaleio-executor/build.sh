#! /usr/bin/env bash

grep -R --exclude-dir vendor --exclude-dir .git --exclude-dir mesos --exclude build.sh TODO ./

GOOS=linux GOARCH=amd64 go build .
