#!/bin/sh
if [ "$#" -ne 1 ]; then
  echo "usage: <migration name>"
  exit 1
fi

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

value=$(printf "%05d-%s.sql" $((1 + $(ls "$SCRIPT_DIR/migrations" | wc -l))) "$1")
echo "-- Empty file" > "$SCRIPT_DIR/migrations/$value"
echo "Successfully made migration" "$value"
