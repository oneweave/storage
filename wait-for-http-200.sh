#!/bin/bash

uri=$1
echo "Waiting for $uri to be ready..."

for i in {1..30}; do
    # Try connecting. Since it is MongoDB port, curl might return a non-200 but connection will succeed
    if curl -s --connect-timeout 2 "$uri" >/dev/null 2>&1 || nc -z localhost 27017 >/dev/null 2>&1 || bash -c 'cat < /dev/null > /dev/tcp/localhost/27017' >/dev/null 2>&1; then
        echo "Ready!"
        exit 0
    fi
    sleep 1
done

echo "Timeout waiting for database"
exit 1
