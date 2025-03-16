#!/bin/sh

set -e

cd ${TMPDIR}

CLI_TMPDIR=${TMPDIR}cli-$(date "+%Y-%m-%d_%H:%M:%S")

echo "creating tmp dir..."
mkdir ${CLI_TMPDIR}
cd ${CLI_TMPDIR}

echo "downloading nightly release..."
curl -sL "https://github.com/kyma-project/cli/releases/download/0.0.0-dev/kyma_$(uname -s)_$(uname -m).tar.gz" -o ${CLI_TMPDIR}/cli.tar.gz

echo "untaring..."
tar -zxvf ${CLI_TMPDIR}/cli.tar.gz kyma

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
