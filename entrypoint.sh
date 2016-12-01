#!/usr/bin/env sh

echo "Creating default logfile"
DIR=$(dirname "${ERNEST_LOG_FILE}")
mkdir -p $DIR
touch $ERNEST_LOG_FILE
mkdir -p $ERNEST_LOG_CONFIG

echo "Starting logger"
/go/bin/logger

