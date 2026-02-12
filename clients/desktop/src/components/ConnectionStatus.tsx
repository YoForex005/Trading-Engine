/**
 * Connection Status Indicator
 * Shows WebSocket connection health with detailed metrics dropdown
 */

import { useState, useRef, useEffect } from 'react';
import { useConnectionHealth, type ConnectionHealth } from '../hooks/useConnectionHealth';
import { Wifi, WifiOff, AlertTriangle, RefreshCw } from 'lucide-react';

interface ConnectionStatusProps {
  wsRef: React.RefObject<WebSocket | null>;
  reconnectCallback?: () => void;
}

export function ConnectionStatus({ wsRef, reconnectCallback }: ConnectionStatusProps) {
  const health = useConnectionHealth({ wsRef, reconnectCallback });
  const [showDropdown, setShowDropdown] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowDropdown(false);
      }
    };

    if (showDropdown) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showDropdown]);

  const { state, lastUpdateAge, metrics, isStale } = health;

  // Determine display based on state
  const getStatusDisplay = () => {
    switch (state) {
      case 'connected':
        return {
          color: 'bg-green-500',
          text: 'Live',
          icon: Wifi,
          iconColor: 'text-green-600',
        };
      case 'connecting':
        return {
          color: 'bg-yellow-500',
          text: `Connecting...`,
          icon: RefreshCw,
          iconColor: 'text-yellow-600 animate-spin',
        };
      case 'disconnected':
        return {
          color: 'bg-red-500',
          text: 'Disconnected',
          icon: WifiOff,
          iconColor: 'text-red-600',
        };
      case 'offline':
        return {
          color: 'bg-gray-500',
          text: 'Offline',
          icon: WifiOff,
          iconColor: 'text-gray-600',
        };
      case 'stale':
        return {
          color: 'bg-orange-500',
          text: 'Stale Data',
          icon: AlertTriangle,
          iconColor: 'text-orange-600',
        };
      default:
        return {
          color: 'bg-gray-500',
          text: 'Unknown',
          icon: WifiOff,
          iconColor: 'text-gray-600',
        };
    }
  };

  const status = getStatusDisplay();
  const StatusIcon = status.icon;

  // Format uptime
  const formatUptime = (seconds: number): string => {
    if (seconds < 60) return `${seconds}s`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
    const hours = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${mins}m`;
  };

  // Format last update
  const formatLastUpdate = (): string => {
    if (lastUpdateAge === 0) return 'Never';
    if (lastUpdateAge < 60) return `${lastUpdateAge}s ago`;
    if (lastUpdateAge < 3600) return `${Math.floor(lastUpdateAge / 60)}m ago`;
    return `${Math.floor(lastUpdateAge / 3600)}h ago`;
  };

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Status Indicator (Clickable) */}
      <button
        onClick={() => setShowDropdown(!showDropdown)}
        className="flex items-center gap-1.5 hover:bg-zinc-300/50 px-2 py-0.5 rounded transition-colors"
        title="Click for connection details"
      >
        {/* Colored Dot */}
        <div className={`w-2 h-2 rounded-full ${status.color} shadow-sm`} />

        {/* Status Text */}
        <span className="font-medium text-black text-[11px]">
          {status.text}
        </span>

        {/* Stale Warning */}
        {isStale && (
          <span className="text-orange-600 text-[10px] font-medium ml-1">
            (Stale)
          </span>
        )}

        {/* Last Update Age */}
        {state === 'connected' && !isStale && (
          <span className="text-zinc-600 text-[10px] ml-1">
            {formatLastUpdate()}
          </span>
        )}

        {/* Reconnect Attempts */}
        {state === 'connecting' && metrics.reconnectAttempts > 0 && (
          <span className="text-yellow-700 text-[10px] ml-1">
            (attempt {metrics.reconnectAttempts})
          </span>
        )}

        {/* Icon */}
        <StatusIcon size={12} className={status.iconColor} />
      </button>

      {/* Dropdown (Detailed Metrics) */}
      {showDropdown && (
        <div className="absolute bottom-full right-0 mb-1 w-64 bg-white border border-zinc-300 rounded shadow-xl z-[100] text-[11px]">
          {/* Header */}
          <div className="bg-zinc-100 border-b border-zinc-300 px-3 py-2">
            <div className="flex items-center gap-2">
              <div className={`w-2.5 h-2.5 rounded-full ${status.color} shadow-sm`} />
              <span className="font-semibold text-black">Connection Status</span>
            </div>
          </div>

          {/* Metrics */}
          <div className="px-3 py-2 space-y-2">
            {/* Connection State */}
            <div className="flex justify-between">
              <span className="text-zinc-600">State:</span>
              <span className="font-medium text-black capitalize">{state}</span>
            </div>

            {/* Server URL */}
            {metrics.serverUrl && (
              <div className="flex justify-between items-start">
                <span className="text-zinc-600">Server:</span>
                <span className="font-medium text-black text-right max-w-[150px] truncate" title={metrics.serverUrl}>
                  {new URL(metrics.serverUrl, window.location.origin).host}
                </span>
              </div>
            )}

            {/* Latency */}
            {metrics.latency > 0 && (
              <div className="flex justify-between">
                <span className="text-zinc-600">Latency:</span>
                <span className={`font-medium ${metrics.latency < 100 ? 'text-green-600' : metrics.latency < 300 ? 'text-yellow-600' : 'text-red-600'}`}>
                  {metrics.latency}ms
                </span>
              </div>
            )}

            {/* Messages Received */}
            <div className="flex justify-between">
              <span className="text-zinc-600">Messages:</span>
              <span className="font-medium text-black">{metrics.messagesReceived.toLocaleString()}</span>
            </div>

            {/* Uptime */}
            {metrics.uptime > 0 && (
              <div className="flex justify-between">
                <span className="text-zinc-600">Uptime:</span>
                <span className="font-medium text-black">{formatUptime(metrics.uptime)}</span>
              </div>
            )}

            {/* Reconnect Attempts */}
            {metrics.reconnectAttempts > 0 && (
              <div className="flex justify-between">
                <span className="text-zinc-600">Reconnect Attempts:</span>
                <span className="font-medium text-orange-600">{metrics.reconnectAttempts}</span>
              </div>
            )}

            {/* Last Update */}
            <div className="flex justify-between">
              <span className="text-zinc-600">Last Update:</span>
              <span className="font-medium text-black">{formatLastUpdate()}</span>
            </div>

            {/* Stale Data Warning */}
            {isStale && (
              <div className="flex items-start gap-2 p-2 bg-orange-50 border border-orange-200 rounded">
                <AlertTriangle size={14} className="text-orange-600 mt-0.5 flex-shrink-0" />
                <span className="text-orange-800 text-[10px]">
                  No data received for over 60 seconds. Connection may be stale.
                </span>
              </div>
            )}
          </div>

          {/* Actions */}
          <div className="border-t border-zinc-300 px-3 py-2">
            <button
              onClick={() => {
                health.reconnect();
                setShowDropdown(false);
              }}
              className="w-full flex items-center justify-center gap-2 px-3 py-1.5 bg-blue-500 hover:bg-blue-600 text-white rounded transition-colors"
              disabled={state === 'connecting'}
            >
              <RefreshCw size={12} className={state === 'connecting' ? 'animate-spin' : ''} />
              <span className="text-[11px] font-medium">
                {state === 'connecting' ? 'Reconnecting...' : 'Reconnect Now'}
              </span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
