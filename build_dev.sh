#!/usr/bin/env bash
echo "1"

echo "2"
LABEL=DEV
echo "3"
docker build -t cooll3r/multi-level-cache:$(cat VERSION)-$LABEL .
echo "4"
