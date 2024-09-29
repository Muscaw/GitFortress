#!/usr/bin/env bash

set -eux pipefail

function build() {
  env GOOS=$1 GOARCH=$2 go build -o ../build/gitfortress-$1-$2 cmd/app/main.go
}


build "darwin" "amd64"
build "darwin" "arm64"
build "linux" "arm"
build "linux" "arm64"
build "linux" "amd64"
