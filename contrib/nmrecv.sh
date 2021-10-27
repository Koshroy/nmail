#!/bin/sh

if [ -z $NNCP_SENDER ]; then
    echo "NNCP_SENDER not provided!"
    exit 1
fi

RECIPIENT="$1"
if [ -z $RECIPIENT ]; then
    echo "No recipient found."
    exit 1
fi

nmail recv | sendmail "$RECIPIENT"
