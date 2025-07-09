#!/usr/bin/env bash

LABEL=DEV
docker build -t cooll3r/multi-level-cache:$(cat VERSION)-$LABEL .
docker push cooll3r/multi-level-cache:$(cat VERSION)-$LABEL

