#!/usr/bin/env bash
set -euo pipefail
LABEL=RELEASE
docker build -t cooll3r/multi-level-cache:$(cat VERSION)-$LABEL .
