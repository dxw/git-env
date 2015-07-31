#!/bin/sh
set -xe

# We don't want to build something that's broken
go vet
errcheck
go test -coverprofile=/tmp/coverage.out
go tool cover -func=/tmp/coverage.out

# Clean
rm -rf build
mkdir -p build/linux64 build/darwin64

# Build
GOARCH=amd64 GOOS=linux  go build -o build/linux64/git-env
GOARCH=amd64 GOOS=darwin go build -o build/darwin64/git-env

# Tar
tar -C build -cjf build/linux64.tar.bz2 linux64
tar -C build -cjf build/darwin64.tar.bz2 darwin64
