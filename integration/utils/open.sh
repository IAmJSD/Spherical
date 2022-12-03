#!/bin/sh
set -e
OPEN_BIN=$(which xdg-open || which open)
if [ -z "$OPEN_BIN" ]; then
  echo "No open command found"
  exit 1
fi
if [ -z "$CI" ]; then
  $OPEN_BIN "$@"
fi
