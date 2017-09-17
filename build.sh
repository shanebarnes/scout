#!/bin/bash

set -e

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export GOPATH="$script_dir"
export GOBIN="${GOPATH}/bin"
#go env

cd "$GOPATH"
mkdir -p "$GOBIN"

which glide &> /dev/null

if [ $? -eq 0 ]; then
    glide install
else
    printf "Downloading and installing packages and dependencies...\n"
    go get ./...
fi

printf "Compiling packages and dependencies...\n"
go build -ldflags -s

exit $?
