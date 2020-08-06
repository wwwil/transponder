#!/usr/bin/env bash

set -o errexit
set -o pipefail

# Use xtrace for debugging.
#set -o xtrace

# Capture arguments then disable unset variables.
FUNC=$1
set -o nounset

REPO_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
BUILD_DIR="${REPO_ROOT}/builds"

VERSION="v0.0.1"
COMMIT=$(git rev-list -1 HEAD)
DATE=$(date -uR)
GOVERSION=$(go version | awk '{print $3 " " $4}')

DOCKER_IMAGE="quay.io/wwwil/transponder"
DOCKER_IMAGE_TAG="${DOCKER_IMAGE}:${VERSION}"

function build() {
    cd "$REPO_ROOT"
    BUILD_GOOS=${1:-$(go env GOOS)}
    BUILD_GOARCH=${2:-$(go env GOARCH)}
    BUILD_OUTPUT="${BUILD_DIR}/transponder-${VERSION}-${BUILD_GOOS}-${BUILD_GOARCH}/transponder"
    LDFLAGS="-X \"github.com/wwwil/transponder/pkg/version.TransponderVersion=${VERSION}\""
    LDFLAGS="${LDFLAGS} -X \"github.com/wwwil/transponder/pkg/version.Platform=${BUILD_GOOS}/${BUILD_GOARCH}\""
    LDFLAGS="${LDFLAGS} -X \"github.com/wwwil/transponder/pkg/version.Commit=${COMMIT}\""
    LDFLAGS="${LDFLAGS} -X \"github.com/wwwil/transponder/pkg/version.BuildDate=${DATE}\""
    LDFLAGS="${LDFLAGS} -X \"github.com/wwwil/transponder/pkg/version.GoVersion=${GOVERSION}\""
    echo "Building Transponder ${VERSION} ${BUILD_GOOS} ${BUILD_GOARCH}."
    GOOS=$BUILD_GOOS GOARCH=$BUILD_GOARCH CGO_ENABLED=0 GO111MODULE=on go build -ldflags "${LDFLAGS}" -o "${BUILD_OUTPUT}" .
}

function build-all() {
    build linux amd64
    build linux arm
    build linux arm64
    build darwin amd64
    build windows amd64
}

function package-all() {
    if [ ! -d "${BUILD_DIR}" ]; then
        echo "No builds to package."
        exit
    fi
    for BUILD in "${BUILD_DIR}"/*; do
        ZIP_NAME=${BUILD_DIR}/$(basename "${BUILD}").zip
        echo "Packaging build ${ZIP_NAME}."
        zip -q "${ZIP_NAME}" "${BUILD}"/*
    done
}

function install() {
    build
    cp "${BUILD_OUTPUT}" "${GOPATH}/bin"
}

function clean() {
    rm -r "${BUILD_DIR}"
}

function docker-build() {
    echo "Building Transponder Docker image ${DOCKER_IMAGE_TAG} for current platform."
    docker build --tag ${DOCKER_IMAGE_TAG} .
	  docker image prune --force --filter label=transponder=docker-build
}

function docker-push() {
    docker-build
    echo "Pushing Docker image ${DOCKER_IMAGE_TAG}. This requires Docker to be logged in with an authorised account."
	  docker push ${DOCKER_IMAGE_TAG}
}

function docker-build-push-all() {
    DOCKER_PLATFORMS="linux/amd64,linux/arm,linux/arm64"
    echo "Building and pushing Transponder Docker image ${DOCKER_IMAGE_TAG} for platforms ${DOCKER_PLATFORMS}. This requires Docker to be logged in with an authorised account."
    docker buildx build --push --tag ${DOCKER_IMAGE_TAG} --platform "${DOCKER_PLATFORMS}" .
	  docker image prune --force --filter label=transponder=docker-build
}

# Check the first argument passed is a function.
if [ "$(type -t $FUNC)" != "function" ]; then
    # If not warn and print the list of functions.
    echo "No target: $FUNC"
    echo "Try: $0 { $(compgen -A function | tr '\n' ' ')}"
    exit 1
fi

# Run the function named by the first argument.
$FUNC
