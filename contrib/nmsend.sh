#!/bin/sh
RECIPIENT="$1"

if [ -z $RECIPIENT ]; then
    echo "No recipient provided!"
    exit 1
fi

nmail send "$RECIPIENT"
