#!/bin/bash
echo "Forcing Backend Restart..."

# Find process ID using port 8080
PID=$(lsof -ti:8080)

if [ -z "$PID" ]; then
    echo "No process found on port 8080."
else
    echo "Killing process $PID..."
    kill -9 $PID
    echo "Killed."
fi
