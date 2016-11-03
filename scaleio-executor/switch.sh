#! /usr/bin/env bash

if [ -f glide.yaml.dev ]; then
    echo "Enabling DEV"
    mv glide.yaml glide.yaml.org
    mv glide.yaml.dev glide.yaml
elif [ -f glide.yaml.org ]; then
    echo "Enabling Build"
    mv glide.yaml glide.yaml.dev
    mv glide.yaml.org glide.yaml
fi