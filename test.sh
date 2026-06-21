#!/bin/bash

set -euo pipefail

declare stopMongoDbAfterTest=''
declare stop_called=''

function stopDockerDependencies() {
    if [ -n "${stopMongoDbAfterTest}" ]; then
        docker compose -f mongodb-compose.yml down
    fi
}

# Start MongoDB if not running
if [ "$(docker ps | grep -c mongo)" == '0' ]; then
    echo "Starting MongoDB for tests"
    docker compose -f mongodb-compose.yml up -d
    ./wait-for-http-200.sh http://localhost:27017/
    
    stopMongoDbAfterTest='true'
fi

stop() {
    exitCode=$1
    
    # prevent re-entrance
    if [ "${stop_called:-}" = "true" ]; then
        return
    fi
    stop_called='true'
    
    stopDockerDependencies;
    exit $exitCode
}

# Capture exit code safely and ensure cleanup on normal or error exit
trap 'rc=$?; stop $rc' EXIT
# Handle Ctrl+C and termination explicitly
trap 'echo "Received SIGINT"; stop 130; exit 130' SIGINT
trap 'echo "Received SIGTERM"; stop 143; exit 143' SIGTERM

# Run tests with MongoDB environment variables
OWV_MONGODB_PASSWORD=password OWV_MONGODB_USERNAME=admin go test -cover -timeout 30s ./...

stop "$?"