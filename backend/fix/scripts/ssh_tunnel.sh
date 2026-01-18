#!/bin/bash
# SSH Tunnel for FIX Connection
# Usage: ./ssh_tunnel.sh user@your-server-ip

if [ -z "$1" ]; then
    echo "Usage: ./ssh_tunnel.sh user@your-server-ip"
    echo ""
    echo "This creates a tunnel so your local machine can reach T4B's FIX server"
    echo "through your VPS/server that doesn't have port blocking."
    echo ""
    echo "After running this, update gateway.go to use:"
    echo '  Host: "127.0.0.1"'
    echo '  Port: 12336'
    exit 1
fi

SERVER=$1
T4B_HOST="23.106.238.138"
T4B_PORT="12336"
LOCAL_PORT="12336"

echo "Creating SSH tunnel..."
echo "  Local:  127.0.0.1:$LOCAL_PORT"
echo "  Remote: $T4B_HOST:$T4B_PORT"
echo "  Via:    $SERVER"
echo ""
echo "Press Ctrl+C to stop the tunnel"
echo ""

ssh -L $LOCAL_PORT:$T4B_HOST:$T4B_PORT $SERVER -N -v
