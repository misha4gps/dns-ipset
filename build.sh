#!/bin/bash

set -e

mkdir -p build

echo "build linux amd64"
CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -o build/dns-ipset-linux-amd64 .
cd build
tar -cvzf dns-ipset-linux-amd64.tar.gz dns-ipset-linux-amd64
cd ../

echo "build linux 386"
CGO_ENABLED=0 GOOS="linux" GOARCH="386" go build -o build/dns-ipset-linux-386 .
cd build
tar -cvzf dns-ipset-linux-386.tar.gz dns-ipset-linux-386
cd ../