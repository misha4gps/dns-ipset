#!/bin/bash

set -e

TARGETS=(
amd64
386
mips
mipsle
mips64
mips64le
arm
)

mkdir -p build


for TARGET in ${TARGETS[@]}; do
  echo "build linux $TARGET"
  # if target starts with mips*, , add argument: GOMIPS=softfloat
  CGO_ENABLED=0 GOOS="linux" GOARCH="$TARGET" go build -trimpath -ldflags="-s -w" -o build/dns-ipset-linux-$TARGET .
  cd build
  tar -cvzf dns-ipset-linux-$TARGET.tar.gz dns-ipset-linux-$TARGET
  cd ../

  if [[ $TARGET == mips* ]]; then
    echo "build linux $TARGET-softfloat"
    CGO_ENABLED=0 GOOS="linux" GOARCH="$TARGET" GOMIPS=softfloat go build -trimpath -ldflags="-s -w" -o build/dns-ipset-linux-$TARGET-softfloat .
    cd build
    tar -cvzf dns-ipset-linux-$TARGET-softfloat.tar.gz dns-ipset-linux-$TARGET-softfloat
    cd ../
  fi
done