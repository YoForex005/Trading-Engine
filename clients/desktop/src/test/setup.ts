import '@testing-library/jest-dom';
import { afterEach, vi, beforeEach } from 'vitest';
import { cleanup } from '@testing-library/react';

// Cleanup after each test
afterEach(() => {
  cleanup();
});

// Mock localStorage
class LocalStorageMock implements Storage {
  private store = new Map<string, string>();

  getItem(key: string): string | null {
    return this.store.get(key) ?? null;
  }

  setItem(key: string, value: string): void {
    this.store.set(key, value);
  }

  removeItem(key: string): void {
    this.store.delete(key);
  }

  clear(): void {
    this.store.clear();
  }

  key(index: number): string | null {
    const keys = Array.from(this.store.keys());
    return keys[index] ?? null;
  }

  get length(): number {
    return this.store.size;
  }
}

// Set up localStorage globally
const localStorageMock = new LocalStorageMock();

beforeEach(() => {
  // Reset localStorage before each test
  localStorageMock.clear();
});

// Use Object.defineProperty to ensure localStorage is available globally
Object.defineProperty(globalThis, 'localStorage', {
  value: localStorageMock,
  writable: true,
  configurable: true,
});

// Also set it on globalThis for Node.js environments
if (typeof window !== 'undefined') {
  Object.defineProperty(window, 'localStorage', {
    value: localStorageMock,
    writable: true,
    configurable: true,
  });
}

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock IntersectionObserver
globalThis.IntersectionObserver = class IntersectionObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  takeRecords() {
    return [];
  }
  unobserve() {}
} as any;

// Mock ResizeObserver
globalThis.ResizeObserver = class ResizeObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  unobserve() {}
} as any;

// Mock WebSocket for tests
// This prevents actual WebSocket connections during tests
type MockWebSocketEventHandler = (event?: any) => void;

class MockWebSocket {
  url: string;
  readyState: number = 0; // CONNECTING
  private handlers: Map<string, MockWebSocketEventHandler[]> = new Map();

  constructor(url: string) {
    this.url = url;
    // Simulate async connection
    setTimeout(() => {
      this.readyState = 1; // OPEN
      this.triggerEvent('open');
    }, 0);
  }

  send(_data: string) {
    // Mock implementation - stores sent data for verification in tests
  }

  close() {
    this.readyState = 3; // CLOSED
    this.triggerEvent('close');
  }

  addEventListener(event: string, handler: MockWebSocketEventHandler) {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, []);
    }
    this.handlers.get(event)!.push(handler);
  }

  removeEventListener(event: string, handler: MockWebSocketEventHandler) {
    const eventHandlers = this.handlers.get(event);
    if (eventHandlers) {
      const index = eventHandlers.indexOf(handler);
      if (index > -1) {
        eventHandlers.splice(index, 1);
      }
    }
  }

  private triggerEvent(event: string, data?: any) {
    const eventHandlers = this.handlers.get(event);
    if (eventHandlers) {
      eventHandlers.forEach((handler) => handler(data));
    }
  }

  // Helper for tests to simulate receiving messages
  simulateMessage(data: any) {
    this.triggerEvent('message', { data: JSON.stringify(data) });
  }
}

// Replace global WebSocket with mock
globalThis.WebSocket = MockWebSocket as any;
