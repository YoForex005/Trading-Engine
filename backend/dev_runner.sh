#!/bin/bash
echo "Starting Backend Dev Runner..."
echo "This script will automatically restart the backend when it crashes or exits."
echo "Use Ctrl+C to stop."

cd "$(dirname "$0")"

while true; do
    echo "----------------------------------------------------------------"
    echo "Running Backend (go run cmd/server/main.go)..."
    echo "----------------------------------------------------------------"
    go run cmd/server/main.go
    
    EXIT_CODE=$?
    echo "Backend exited with code $EXIT_CODE."
    
    if [ $EXIT_CODE -ne 0 ]; then
        echo "Crashed! Restarting in 3 seconds..."
        sleep 3
    else
        echo "Clean exit. Restarting immediately..."
        sleep 1
    fi
done
