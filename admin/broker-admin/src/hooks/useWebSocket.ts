import { useState, useEffect, useRef, useCallback } from 'react';

interface MarketTick {
    type: string;
    symbol: string;
    bid: number;
    ask: number;
    spread: number;
    timestamp: number;
    lp: string;
}

interface UseWebSocketOptions {
    url: string;
    token?: string;
    onMessage?: (tick: MarketTick) => void;
    reconnectInterval?: number;
}

/**
 * Get JWT token from localStorage
 */
function getAuthToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('admin_token') ||
           localStorage.getItem('rtx_token') ||
           localStorage.getItem('jwt_token') ||
           null;
}

export function useWebSocket({ url, token, onMessage, reconnectInterval = 3000 }: UseWebSocketOptions) {
    const [isConnected, setIsConnected] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    const connect = useCallback(() => {
        // Build URL with token - try param first, then localStorage
        const authToken = token || getAuthToken();
        const wsUrl = authToken ? `${url}?token=${encodeURIComponent(authToken)}` : url;

        try {
            const ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                setIsConnected(true);
                setError(null);
            };

            ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data) as MarketTick;
                    if (data.type === 'tick' && onMessage) {
                        onMessage(data);
                    }
                } catch (e) {
                    console.error('[WS] Failed to parse message:', e);
                }
            };

            ws.onclose = () => {
                setIsConnected(false);
                wsRef.current = null;

                // Auto-reconnect
                if (reconnectInterval > 0) {
                    reconnectTimeoutRef.current = setTimeout(() => {
                        connect();
                    }, reconnectInterval);
                }
            };

            ws.onerror = (e) => {
                console.error('[WS] Error:', e);
                setError('WebSocket connection failed');
            };

            wsRef.current = ws;
        } catch (e) {
            console.error('[WS] Failed to create WebSocket:', e);
            setError('Failed to connect');
        }
    }, [url, token, onMessage, reconnectInterval]);

    const disconnect = useCallback(() => {
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
        }
        if (wsRef.current) {
            wsRef.current.close();
            wsRef.current = null;
        }
        setIsConnected(false);
    }, []);

    useEffect(() => {
        connect();
        return () => disconnect();
    }, [connect, disconnect]);

    return { isConnected, error, disconnect, reconnect: connect };
}

export type { MarketTick };
