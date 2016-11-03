#! /usr/bin/env bash

if [ "$1" == "which" ]; then
    if [ -f glide.yaml.dev ]; then
        echo "Building PROD"
    elif [ -f glide.yaml.org ]; then
        echo "Building DEV"
    fi
elif [ -f glide.yaml.dev ]; then
    echo "Enabling DEV"
    mv glide.yaml glide.yaml.org
    mv glide.yaml.dev glide.yaml
elif [ -f glide.yaml.org ]; then
    echo "Enabling PROD"
    mv glide.yaml glide.yaml.dev
    mv glide.yaml.org glide.yaml
fi