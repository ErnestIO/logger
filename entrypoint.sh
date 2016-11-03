#!/usr/bin/env sh

echo "Creating default logfile"
DIR=$(dirname "${ERNEST_LOG_FILE}")
mkdir -p $DIR
touch $ERNEST_LOG_FILE

echo "Starting logger"
/go/bin/logger

