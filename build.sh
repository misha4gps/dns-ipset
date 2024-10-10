#!/bin/bash

mkdir -p build

echo "build linux amd64"
CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -o build/dns-ipset-linux-amd64 .
echo "build linux 386"
CGO_ENABLED=0 GOOS="linux" GOARCH="386" go build -o build/dns-ipset-linux-386 .