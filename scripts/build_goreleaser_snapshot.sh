#!/bin/bash
#
# Allows manual building of packages via goreleaser.
# Allows building of dirty git state and will not push results to github.

# make script executable independent of path
WASP_ROOT_DIR="$( cd -- "$(dirname "$0")/.." >/dev/null 2>&1 ; pwd -P )"

GORELEASER_IMAGE=iotaledger/goreleaser-cgo-cross-compiler:1.20.2
REPO_PATH="/build"

docker pull "${GORELEASER_IMAGE}"
docker run --rm --privileged -v "${WASP_ROOT_DIR}":"${REPO_PATH}" -w "${REPO_PATH}" "${GORELEASER_IMAGE}" goreleaser --clean --snapshot --skip-publish -f ./tools/wasp-cli/.goreleaser.yml
