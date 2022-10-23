#!/bin/sh
if [ "$#" -ne 1 ]; then
  echo "usage: <migration name>"
  exit 1
fi

value=$(printf "%05d-%s.sql" $((1 + $(ls migrations | wc -l))) "$1")
echo "-- Empty file" > migrations/"$value"
echo "Successfully made migration" "$value"
