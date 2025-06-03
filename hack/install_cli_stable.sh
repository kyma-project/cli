#!/bin/sh

set -e

TMPDIR=${TMPDIR:-/tmp}
CLI_TMPDIR=${TMPDIR}/cli-$(date "+%Y-%m-%d_%H:%M:%S")

echo "creating tmp dir..."
mkdir -p ${CLI_TMPDIR}

VERSION=$(curl -sL https://api.github.com/repos/kyma-project/cli/releases/latest | jq -r '.tag_name')

DIST=$(uname -s)
ARCH=$(uname -m)
if [ "${ARCH}" = "amd64" ]; then
    ARCH="x86_64"
elif [ "${ARCH}" = "aarch64" ]; then
    ARCH="arm64"
fi

echo "downloading ${VERSION} release..."
curl -sL "https://github.com/kyma-project/cli/releases/download/${VERSION}/kyma_${DIST}_${ARCH}.tar.gz" -o ${CLI_TMPDIR}/cli.tar.gz

echo "untaring..."
tar -zxvf ${CLI_TMPDIR}/cli.tar.gz --directory ${CLI_TMPDIR} kyma

set +e

echo "moving kyma binary to the /usr/local/bin directory..."
cp ${CLI_TMPDIR}/kyma /usr/local/bin/kyma

if [ $? -gt 0 ]; then
    set -e

    echo "failed to copy, trying with sudo..."
    sudo cp ${CLI_TMPDIR}/kyma /usr/local/bin/kyma
fi

echo "removing tmp dir..."
rm -r ${CLI_TMPDIR}

echo "done"
