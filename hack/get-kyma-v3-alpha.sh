#!/bin/sh
mkdir tmp
cd tmp

OS="$(uname -s)" # e.g Darwin
ARCH="$(uname -m)" # e.g. arm64

curl -Lo kymav3.tar.gz https://github.com/kyma-project/cli/releases/download/v0.0.0-dev/kyma_${OS}_${ARCH}.tar.gz


tar zxvf kymav3.tar.gz 
mkdir -p ../bin
cp kyma ../bin/kyma@v3

cd ..
rm -rfd tmp