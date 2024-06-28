#!/bin/sh
mkdir tmp
cd tmp

curl -Lo kymav3.tar.gz https://github.com/kyma-project/cli/releases/download/v0.0.0-dev/kyma_Darwin_x86_64.tar.gz # kyma-linux, kyma-linux-arm, kyma.exe, or kyma-arm.exe


tar zxvf kymav3.tar.gz 
cp kyma ../bin/kyma@v3
cp kyma /usr/local/bin/kyma@v3

cd ..
rm -rfd tmp