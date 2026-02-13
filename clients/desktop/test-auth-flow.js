#!/usr/bin/env node

/**
 * Authentication Flow Test Script
 * Tests the login and authentication flow
 */

const API_URL = process.env.VITE_API_URL || 'http://localhost:7999';

async function testLoginFlow() {
  console.log('🔒 Testing Authentication Flow\n');
  console.log(`API URL: ${API_URL}\n`);

  // Test 1: Invalid Credentials
  console.log('Test 1: Invalid Credentials');
  try {
    const res = await fetch(`${API_URL}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'invalid', password: 'wrong' })
    });

    if (res.status === 401) {
      console.log('✅ PASS: 401 returned for invalid credentials\n');
    } else {
      console.log(`❌ FAIL: Expected 401, got ${res.status}\n`);
    }
  } catch (err) {
    console.log(`⚠️  ERROR: ${err.message}\n`);
  }

  // Test 2: Valid Credentials
  console.log('Test 2: Valid Credentials (username: 1, password: password)');
  try {
    const res = await fetch(`${API_URL}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: '1', password: 'password' })
    });

    if (res.status === 200) {
      const data = await res.json();

      if (data.token) {
        console.log('✅ PASS: Login successful, token received');
        console.log(`   Token: ${data.token.substring(0, 20)}...`);

        if (data.user) {
          console.log(`   User ID: ${data.user.id}`);
          console.log(`   Username: ${data.user.username}`);
          console.log(`   Role: ${data.user.role}`);
        }

        // Test 3: Token Validation
        console.log('\nTest 3: Token Validation');
        try {
          const accountRes = await fetch(`${API_URL}/api/account/summary?accountId=1`, {
            headers: {
              'Authorization': `Bearer ${data.token}`,
              'Content-Type': 'application/json'
            }
          });

          if (accountRes.ok) {
            console.log('✅ PASS: API request with token successful');
          } else if (accountRes.status === 401) {
            console.log('❌ FAIL: Token not accepted by API');
          } else {
            console.log(`⚠️  Unexpected status: ${accountRes.status}`);
          }
        } catch (err) {
          console.log(`⚠️  ERROR: ${err.message}`);
        }

        console.log('\n');
      } else {
        console.log('❌ FAIL: No token in response\n');
      }
    } else {
      console.log(`❌ FAIL: Expected 200, got ${res.status}\n`);
    }
  } catch (err) {
    console.log(`⚠️  ERROR: ${err.message}\n`);
  }

  // Test 4: Server Health
  console.log('Test 4: Server Health Check');
  try {
    const res = await fetch(`${API_URL}/api/config`);
    if (res.ok) {
      const config = await res.json();
      console.log('✅ PASS: Server is running');
      console.log(`   Broker: ${config.brokerName || 'Unknown'}`);
      console.log(`   Execution Mode: ${config.executionMode || 'Unknown'}`);
    } else {
      console.log(`⚠️  Server responded with ${res.status}`);
    }
  } catch (err) {
    console.log(`❌ FAIL: Server not reachable - ${err.message}`);
  }

  console.log('\n📊 Test Summary');
  console.log('================');
  console.log('Run this script to verify the backend is running correctly.');
  console.log('If all tests pass, the frontend authentication should work.\n');
  console.log('Frontend Checklist:');
  console.log('  1. Login form accepts credentials');
  console.log('  2. Token is stored in localStorage');
  console.log('  3. WebSocket connects with token');
  console.log('  4. API requests include Authorization header');
  console.log('  5. Terminal loads after login\n');
}

// Run tests
testLoginFlow().catch(err => {
  console.error('Fatal error:', err);
  process.exit(1);
});
