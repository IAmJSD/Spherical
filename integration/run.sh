#!/bin/sh
set -e
INTEGRATION_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$INTEGRATION_DIR/.."
docker build . -t spherical_build --build-arg ENABLE_CGO=0 -f Dockerfile.spherical
cd "$INTEGRATION_DIR"
trap "docker image rm spherical_build" EXIT
docker build . -t spherical_integration
docker image rm spherical_build
trap "docker compose down --volumes && docker image rm spherical_integration && ./utils/open.sh playwright-report/index.html" EXIT
mkdir -p test-results
mkdir -p playwright-report
mkdir -p playwright
docker compose up -d
docker attach integration-app-1
