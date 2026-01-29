/**
 * Analytics WebSocket Client
 *
 * Production-ready TypeScript client for consuming real-time analytics updates
 * Features: Auto-reconnect, subscription management, message batching handling
 */

export type Severity = 'info' | 'warning' | 'error' | 'critical'
export type RiskLevel = 'low' | 'medium' | 'high' | 'critical'
export type LPStatus = 'connected' | 'disconnected' | 'degraded'

export type RoutingMetrics = {
  timestamp: string
  symbol: string
  side: string
  volume: number
  routingDecision: 'BBOOK' | 'ABOOK'
  lpSelected?: string
  executionTimeMs: number
  spread: number
  slippage?: number
}

export type LPPerformanceMetrics = {
  timestamp: string
  lpName: string
  status: LPStatus
  avgSpread: number
  executionQuality: number
  latencyMs: number
  quotesPerSecond: number
  rejectRate: number
  uptime: number
}

export type ExposureUpdate = {
  timestamp: string
  totalExposure: number
  bySymbol: Record<string, number>
  byLP: Record<string, number>
  netExposure: number
  exposureLimit: number
  utilizationPct: number
  riskLevel: RiskLevel
}

export type Alert = {
  id: string
  timestamp: string
  severity: Severity
  category: 'exposure' | 'lp' | 'routing' | 'system'
  title: string
  message: string
  source?: string
  actionItems?: string[]
}

export type AnalyticsMessage = {
  type: string
  timestamp: string
  data: RoutingMetrics | LPPerformanceMetrics | ExposureUpdate | Alert
}

export type SubscriptionMessage = {
  action: 'subscribe' | 'unsubscribe'
  channels: string[]
}

export type AnalyticsEventHandlers = {
  onRoutingMetrics?: (data: RoutingMetrics) => void
  onLPPerformance?: (data: LPPerformanceMetrics) => void
  onExposureUpdate?: (data: ExposureUpdate) => void
  onAlert?: (data: Alert) => void
  onConnect?: () => void
  onDisconnect?: () => void
  onError?: (error: Error) => void
  onReconnecting?: (attempt: number) => void
}

export type AnalyticsWebSocketConfig = {
  url: string
  token: string
  channels?: string[]
  autoReconnect?: boolean
  maxReconnectAttempts?: number
  reconnectDelay?: number
  debug?: boolean
}

export class AnalyticsWebSocket {
  private ws: WebSocket | null = null
  private config: Required<AnalyticsWebSocketConfig>
  private handlers: AnalyticsEventHandlers
  private reconnectAttempts = 0
  private reconnectTimer: NodeJS.Timeout | null = null
  private isManualClose = false
  private subscribedChannels = new Set<string>()

  constructor(config: AnalyticsWebSocketConfig, handlers: AnalyticsEventHandlers = {}) {
    this.config = {
      url: config.url,
      token: config.token,
      channels: config.channels || [],
      autoReconnect: config.autoReconnect !== false,
      maxReconnectAttempts: config.maxReconnectAttempts || 10,
      reconnectDelay: config.reconnectDelay || 3000,
      debug: config.debug || false,
    }
    this.handlers = handlers
  }

  connect(): void {
    this.isManualClose = false
    const wsUrl = `${this.config.url}?token=${this.config.token}`

    this.log('Connecting to analytics WebSocket...')

    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      this.log('Connected to analytics WebSocket')
      this.reconnectAttempts = 0

      // Subscribe to initial channels
      if (this.config.channels.length > 0) {
        this.subscribe(this.config.channels)
      }

      this.handlers.onConnect?.()
    }

    this.ws.onmessage = (event) => {
      try {
        const messages: AnalyticsMessage[] = JSON.parse(event.data)

        // Handle batched messages
        messages.forEach((msg) => this.handleMessage(msg))
      } catch (error) {
        this.logError('Failed to parse message:', error)
        this.handlers.onError?.(error as Error)
      }
    }

    this.ws.onerror = (error) => {
      this.logError('WebSocket error:', error)
      this.handlers.onError?.(new Error('WebSocket error'))
    }

    this.ws.onclose = () => {
      this.log('WebSocket connection closed')
      this.handlers.onDisconnect?.()

      // Auto-reconnect if not manually closed
      if (!this.isManualClose && this.config.autoReconnect) {
        this.attemptReconnect()
      }
    }
  }

  disconnect(): void {
    this.isManualClose = true
    this.clearReconnectTimer()

    if (this.ws) {
      this.ws.close()
      this.ws = null
    }

    this.subscribedChannels.clear()
    this.log('Disconnected from analytics WebSocket')
  }

  subscribe(channels: string[]): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.logError('Cannot subscribe: WebSocket not connected')
      return
    }

    const msg: SubscriptionMessage = {
      action: 'subscribe',
      channels,
    }

    this.ws.send(JSON.stringify(msg))
    channels.forEach(ch => this.subscribedChannels.add(ch))

    this.log(`Subscribed to channels: ${channels.join(', ')}`)
  }

  unsubscribe(channels: string[]): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.logError('Cannot unsubscribe: WebSocket not connected')
      return
    }

    const msg: SubscriptionMessage = {
      action: 'unsubscribe',
      channels,
    }

    this.ws.send(JSON.stringify(msg))
    channels.forEach(ch => this.subscribedChannels.delete(ch))

    this.log(`Unsubscribed from channels: ${channels.join(', ')}`)
  }

  getSubscribedChannels(): string[] {
    return Array.from(this.subscribedChannels)
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }

  private handleMessage(msg: AnalyticsMessage): void {
    this.log(`Received ${msg.type}:`, msg.data)

    switch (msg.type) {
      case 'routing-decision':
        this.handlers.onRoutingMetrics?.(msg.data as RoutingMetrics)
        break

      case 'lp-metrics':
        this.handlers.onLPPerformance?.(msg.data as LPPerformanceMetrics)
        break

      case 'exposure-change':
        this.handlers.onExposureUpdate?.(msg.data as ExposureUpdate)
        break

      case 'alert':
        this.handlers.onAlert?.(msg.data as Alert)
        break

      default:
        this.log(`Unknown message type: ${msg.type}`)
    }
  }

  private attemptReconnect(): void {
    this.clearReconnectTimer()

    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      this.logError('Max reconnection attempts reached')
      this.handlers.onError?.(new Error('Max reconnection attempts reached'))
      return
    }

    this.reconnectAttempts++
    this.log(`Reconnecting in ${this.config.reconnectDelay}ms (attempt ${this.reconnectAttempts}/${this.config.maxReconnectAttempts})`)

    this.handlers.onReconnecting?.(this.reconnectAttempts)

    this.reconnectTimer = setTimeout(() => {
      this.connect()
    }, this.config.reconnectDelay)
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }

  private log(...args: unknown[]): void {
    if (this.config.debug) {
      console.log('[AnalyticsWS]', ...args)
    }
  }

  private logError(...args: unknown[]): void {
    console.error('[AnalyticsWS]', ...args)
  }
}

/**
 * Example usage:
 */
export function exampleUsage() {
  const client = new AnalyticsWebSocket(
    {
      url: 'ws://localhost:7999/ws/analytics',
      token: 'your-jwt-token-here',
      channels: ['routing-metrics', 'lp-performance', 'exposure-updates', 'alerts'],
      autoReconnect: true,
      debug: true,
    },
    {
      onConnect: () => {
        console.log('Connected to analytics WebSocket!')
      },

      onRoutingMetrics: (data) => {
        console.log(`[ROUTING] ${data.symbol} ${data.side} → ${data.routingDecision}`)
        console.log(`  Volume: ${data.volume}, LP: ${data.lpSelected}, Latency: ${data.executionTimeMs}ms`)
      },

      onLPPerformance: (data) => {
        console.log(`[LP] ${data.lpName} - ${data.status}`)
        console.log(`  Latency: ${data.latencyMs}ms, QPS: ${data.quotesPerSecond}`)
        console.log(`  Quality: ${(data.executionQuality * 100).toFixed(1)}%, Uptime: ${data.uptime}%`)
      },

      onExposureUpdate: (data) => {
        console.log(`[EXPOSURE] Total: $${data.totalExposure.toLocaleString()}`)
        console.log(`  Utilization: ${data.utilizationPct.toFixed(1)}% (${data.riskLevel})`)
        console.log(`  By Symbol:`, data.bySymbol)
      },

      onAlert: (data) => {
        const emoji = {
          info: 'ℹ️',
          warning: '⚠️',
          error: '❌',
          critical: '🚨',
        }[data.severity]

        console.log(`${emoji} [${data.severity.toUpperCase()}] ${data.title}`)
        console.log(`  ${data.message}`)
        if (data.actionItems?.length) {
          console.log(`  Actions:`)
          data.actionItems.forEach(action => console.log(`    - ${action}`))
        }
      },

      onDisconnect: () => {
        console.log('Disconnected from analytics WebSocket')
      },

      onReconnecting: (attempt) => {
        console.log(`Reconnecting... (attempt ${attempt})`)
      },

      onError: (error) => {
        console.error('WebSocket error:', error)
      },
    }
  )

  // Connect
  client.connect()

  // Later: Subscribe to additional channels
  setTimeout(() => {
    client.subscribe(['alerts'])
  }, 5000)

  // Later: Unsubscribe from channels
  setTimeout(() => {
    client.unsubscribe(['routing-metrics'])
  }, 10000)

  // Later: Disconnect
  setTimeout(() => {
    client.disconnect()
  }, 60000)

  return client
}
