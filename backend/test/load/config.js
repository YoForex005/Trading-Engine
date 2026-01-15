// Shared configuration for load tests

export const config = {
  // Base URLs (override with environment variables)
  wsURL: __ENV.WS_URL || 'ws://localhost:8080/ws',
  apiURL: __ENV.API_URL || 'http://localhost:8080',

  // Performance targets
  targets: {
    maxConcurrentConnections: 100,
    maxOrdersPerSecond: 50,
    p95TickLatency: 500,      // milliseconds
    p95OrderLatency: 200,     // milliseconds
  },

  // Test user credentials
  testUser: {
    username: __ENV.TEST_USER || 'loadtest',
    password: __ENV.TEST_PASS || 'loadtest123',
  }
};
