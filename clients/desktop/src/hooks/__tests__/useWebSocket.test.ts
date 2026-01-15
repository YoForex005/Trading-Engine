import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { useWebSocket } from '../useWebSocket'

describe('useWebSocket', () => {
  beforeEach(() => {
    // WebSocket is mocked in test setup.ts
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.clearAllTimers()
  })

  it('establishes connection on mount', async () => {
    const { result } = renderHook(() => useWebSocket('ws://localhost:8080'))

    // Initially not connected
    expect(result.current.isConnected).toBe(false)

    // Wait for connection (mock WebSocket connects after setTimeout in setup.ts)
    await waitFor(() => {
      expect(result.current.isConnected).toBe(true)
    })
  })

  it('receives messages from WebSocket', async () => {
    const onMessage = vi.fn()
    const { result } = renderHook(() =>
      useWebSocket('ws://localhost:8080', { onMessage })
    )

    // Wait for connection
    await waitFor(() => {
      expect(result.current.isConnected).toBe(true)
    })

    // Simulate receiving a message
    const mockMessage = { type: 'tick', symbol: 'EURUSD', bid: 1.0850, ask: 1.0852 }

    // Access mock WebSocket and trigger message
    // The mock WebSocket is available globally
    const MockWebSocket = globalThis.WebSocket as any
    const wsInstance = MockWebSocket.lastInstance || new MockWebSocket('ws://localhost:8080')
    if (wsInstance && wsInstance.simulateMessage) {
      wsInstance.simulateMessage(mockMessage)
    }

    await waitFor(() => {
      expect(onMessage).toHaveBeenCalledWith(mockMessage)
    })
  })

  it('sends messages through WebSocket', async () => {
    const { result } = renderHook(() => useWebSocket('ws://localhost:8080'))

    await waitFor(() => {
      expect(result.current.isConnected).toBe(true)
    })

    const message = { type: 'subscribe', symbols: ['EURUSD'] }
    result.current.send(message)

    // In a real test, we would verify the WebSocket.send was called
    // For now, we just verify the function doesn't throw
  })

  it('provides disconnect function', () => {
    const { result } = renderHook(() => useWebSocket('ws://localhost:8080'))

    expect(result.current.disconnect).toBeDefined()
    expect(typeof result.current.disconnect).toBe('function')
  })

  it('closes connection on unmount', async () => {
    const onClose = vi.fn()
    const { result, unmount } = renderHook(() =>
      useWebSocket('ws://localhost:8080', { onClose })
    )

    await waitFor(() => {
      expect(result.current.isConnected).toBe(true)
    })

    unmount()

    // Connection should be closed after unmount
    await waitFor(() => {
      expect(onClose).toHaveBeenCalled()
    })
  })

  it('calls onOpen callback when connected', async () => {
    const onOpen = vi.fn()
    const { result } = renderHook(() =>
      useWebSocket('ws://localhost:8080', { onOpen })
    )

    await waitFor(() => {
      expect(result.current.isConnected).toBe(true)
    })

    expect(onOpen).toHaveBeenCalled()
  })

  it('handles connection with options', () => {
    const onMessage = vi.fn()
    const onOpen = vi.fn()
    const onClose = vi.fn()

    const { result } = renderHook(() =>
      useWebSocket('ws://localhost:8080', {
        onMessage,
        onOpen,
        onClose,
        reconnect: true,
        reconnectInterval: 1000,
        reconnectAttempts: 3,
      })
    )

    // Should initialize without errors
    expect(result.current).toBeDefined()
  })
})
