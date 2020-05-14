#!/usr/bin/env bash

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CHANGELOG_GENERATOR="/Users/i531200/go/src/github.com/kyma-project/test-infra/prow/scripts/changelog-generator.sh"
delete_stable_tag() {
    if [ $(git tag -l stable) ]; then
        git tag -d stable
    fi
}

main() {
    # git remote add origin git@github.com:kyma-project/cli.git
    # locally delete the stable tag in order to have all the change logs listed under the new release version
    delete_stable_tag
    # bash "${CHANGELOG_GENERATOR}"
    curl -sL https://git.io/goreleaser | VERSION=v0.118.2 bash  -s --  --release-notes .changelog/release-changelog.md --skip-publish
}

main