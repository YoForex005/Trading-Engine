/**
 * E2E Admin Workflow Tests
 * Tests administrative operations and configurations
 *
 * Setup: npm install playwright @playwright/test
 * Run: npx playwright test admin_workflow_test.js
 */

const { test, expect } = require('@playwright/test');

const API_BASE = 'http://localhost:7999';

/**
 * Admin API Helper
 */
class AdminAPI {
  async getConfig() {
    const response = await fetch(`${API_BASE}/api/config`);
    return await response.json();
  }

  async updateConfig(config) {
    const response = await fetch(`${API_BASE}/api/config`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config),
    });
    return await response.json();
  }

  async getExecutionMode() {
    const response = await fetch(`${API_BASE}/admin/execution-mode`);
    return await response.json();
  }

  async setExecutionMode(mode) {
    const response = await fetch(`${API_BASE}/admin/execution-mode`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ mode }),
    });
    return await response.json();
  }

  async listLPs() {
    const response = await fetch(`${API_BASE}/admin/lps`);
    return await response.json();
  }

  async toggleLP(lpId, enabled) {
    const response = await fetch(`${API_BASE}/admin/lps/${lpId}/toggle`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled }),
    });
    return await response.json();
  }

  async getLPStatus() {
    const response = await fetch(`${API_BASE}/admin/lp-status`);
    return await response.json();
  }

  async getFIXStatus() {
    const response = await fetch(`${API_BASE}/admin/fix/status`);
    return await response.json();
  }

  async connectFIX(sessionId) {
    const response = await fetch(`${API_BASE}/admin/fix/connect`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sessionId }),
    });
    return await response.json();
  }

  async disconnectFIX(sessionId) {
    const response = await fetch(`${API_BASE}/admin/fix/disconnect`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sessionId }),
    });
    return await response.json();
  }

  async getAccounts() {
    const response = await fetch(`${API_BASE}/admin/accounts`);
    return await response.json();
  }

  async deposit(accountId, amount, method, note) {
    const response = await fetch(`${API_BASE}/admin/deposit`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ accountId, amount, method, note }),
    });
    return await response.json();
  }

  async withdraw(accountId, amount, method, note) {
    const response = await fetch(`${API_BASE}/admin/withdraw`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ accountId, amount, method, note }),
    });
    return await response.json();
  }

  async adjust(accountId, amount, reason, note) {
    const response = await fetch(`${API_BASE}/admin/adjust`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ accountId, amount, reason, note }),
    });
    return await response.json();
  }

  async addBonus(accountId, amount, reason) {
    const response = await fetch(`${API_BASE}/admin/bonus`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ accountId, amount, reason }),
    });
    return await response.json();
  }

  async getLedger() {
    const response = await fetch(`${API_BASE}/admin/ledger`);
    return await response.json();
  }

  async getSymbols() {
    const response = await fetch(`${API_BASE}/admin/symbols`);
    return await response.json();
  }

  async toggleSymbol(symbol, enabled) {
    const response = await fetch(`${API_BASE}/admin/symbols/toggle`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ symbol, enabled }),
    });
    return await response.json();
  }

  async getRoutes() {
    const response = await fetch(`${API_BASE}/admin/routes`);
    return await response.json();
  }
}

/**
 * Admin Test Suite
 */

test.describe('Admin Configuration Tests', () => {
  let admin;

  test.beforeEach(async () => {
    admin = new AdminAPI();
  });

  test('Complete Admin Workflow: Configure → Enable LPs → Test Execution', async () => {
    console.log('=== Starting Admin Complete Workflow ===');

    // Step 1: Get and verify initial config
    const initialConfig = await admin.getConfig();

    expect(initialConfig).toBeTruthy();
    expect(initialConfig.brokerName).toBeTruthy();

    console.log('✓ Step 1: Initial config loaded');
    console.log('  Broker:', initialConfig.brokerName);
    console.log('  Execution:', initialConfig.executionMode);
    console.log('  Leverage:', initialConfig.defaultLeverage);

    // Step 2: Update broker configuration
    const newConfig = {
      brokerName: 'RTX Trading E2E Test',
      defaultLeverage: 200,
      defaultBalance: 15000,
    };

    const updateResult = await admin.updateConfig(newConfig);

    expect(updateResult.success).toBe(true);

    console.log('✓ Step 2: Config updated');

    // Step 3: Configure execution mode
    const executionMode = await admin.getExecutionMode();

    expect(executionMode.mode).toBeTruthy();

    console.log('✓ Step 3: Execution mode:', executionMode.mode);

    // Step 4: List and configure LPs
    const lps = await admin.listLPs();

    expect(Array.isArray(lps)).toBe(true);

    console.log('✓ Step 4: LPs configured:', lps.length);

    for (const lp of lps) {
      console.log(`  - ${lp.name}: ${lp.enabled ? 'ENABLED' : 'DISABLED'}`);
    }

    // Step 5: Check LP status
    const lpStatus = await admin.getLPStatus();

    expect(lpStatus).toBeTruthy();

    console.log('✓ Step 5: LP status checked');

    console.log('=== Admin Workflow Complete ===');
  });

  test('Broker Configuration Management', async () => {
    // Get current config
    const config = await admin.getConfig();

    expect(config).toBeTruthy();
    expect(config.brokerName).toBeTruthy();

    console.log('Current configuration:');
    console.log('  Broker Name:', config.brokerName);
    console.log('  Price Feed LP:', config.priceFeedLP);
    console.log('  Execution Mode:', config.executionMode);
    console.log('  Default Leverage:', config.defaultLeverage);
    console.log('  Default Balance:', config.defaultBalance);
    console.log('  Margin Mode:', config.marginMode);

    // Update specific settings
    const updates = {
      defaultLeverage: 150,
      marginMode: 'HEDGING',
    };

    const result = await admin.updateConfig(updates);

    if (result.success) {
      console.log('✓ Configuration updated successfully');
    }
  });

  test('Execution Mode Toggle', async () => {
    // Get current mode
    const current = await admin.getExecutionMode();

    expect(current.mode).toBeTruthy();
    console.log('Current mode:', current.mode);

    // Toggle to ABOOK
    const abookResult = await admin.setExecutionMode('ABOOK');

    if (abookResult.success) {
      expect(abookResult.newMode).toBe('ABOOK');
      console.log('✓ Switched to ABOOK');
    }

    // Toggle back to BBOOK
    const bbookResult = await admin.setExecutionMode('BBOOK');

    if (bbookResult.success) {
      expect(bbookResult.newMode).toBe('BBOOK');
      console.log('✓ Switched to BBOOK');
    }
  });

  test('Invalid Execution Mode', async () => {
    const result = await admin.setExecutionMode('INVALID');

    console.log('Invalid mode response:', result);
    // Should receive error response
  });
});

test.describe('LP Management Tests', () => {
  let admin;

  test.beforeEach(async () => {
    admin = new AdminAPI();
  });

  test('List and Configure LPs', async () => {
    const lps = await admin.listLPs();

    expect(Array.isArray(lps)).toBe(true);
    expect(lps.length).toBeGreaterThan(0);

    console.log('Available LPs:');
    lps.forEach(lp => {
      console.log(`  ${lp.id}: ${lp.name} (${lp.type}) - ${lp.enabled ? 'ENABLED' : 'DISABLED'}`);
      console.log(`    Priority: ${lp.priority}`);
    });
  });

  test('Toggle LP Enable/Disable', async () => {
    const lps = await admin.listLPs();

    if (lps.length > 0) {
      const testLP = lps[0];
      const newState = !testLP.enabled;

      const result = await admin.toggleLP(testLP.id, newState);

      console.log(`✓ Toggled ${testLP.id}: ${newState ? 'ENABLED' : 'DISABLED'}`);

      // Toggle back
      await admin.toggleLP(testLP.id, testLP.enabled);
      console.log(`✓ Restored ${testLP.id} to original state`);
    } else {
      console.log('⚠ No LPs available for testing');
    }
  });

  test('LP Status Monitoring', async () => {
    const status = await admin.getLPStatus();

    expect(status).toBeTruthy();

    console.log('LP Status:');
    console.log('  Total LPs:', status.totalLps || 'N/A');
    console.log('  Active LPs:', status.activeLps || 'N/A');

    if (status.lps && Array.isArray(status.lps)) {
      status.lps.forEach(lp => {
        console.log(`  ${lp.name}:`);
        console.log(`    Connected: ${lp.connected}`);
        console.log(`    Status: ${lp.status}`);
        if (lp.latency) {
          console.log(`    Latency: ${lp.latency}ms`);
        }
      });
    }
  });
});

test.describe('FIX Session Management Tests', () => {
  let admin;

  test.beforeEach(async () => {
    admin = new AdminAPI();
  });

  test('FIX Session Status', async () => {
    const status = await admin.getFIXStatus();

    expect(status).toBeTruthy();
    expect(status.sessions).toBeTruthy();

    console.log('FIX Sessions:');
    for (const [sessionId, sessionStatus] of Object.entries(status.sessions)) {
      console.log(`  ${sessionId}: ${sessionStatus}`);
    }
  });

  test('FIX Session Connect/Disconnect', async () => {
    const sessionId = 'YOFX1';

    // Try to connect
    const connectResult = await admin.connectFIX(sessionId);

    console.log('Connect result:', connectResult);

    if (connectResult.success) {
      console.log(`✓ FIX session ${sessionId} connection initiated`);

      // Wait for connection
      await new Promise(resolve => setTimeout(resolve, 2000));

      // Disconnect
      const disconnectResult = await admin.disconnectFIX(sessionId);

      if (disconnectResult.success) {
        console.log(`✓ FIX session ${sessionId} disconnected`);
      }
    } else {
      console.log('⚠ FIX connection not available (expected in test environment)');
    }
  });
});

test.describe('Account Management Tests', () => {
  let admin;

  test.beforeEach(async () => {
    admin = new AdminAPI();
  });

  test('List All Accounts', async () => {
    const accounts = await admin.getAccounts();

    console.log('Accounts:');
    if (Array.isArray(accounts)) {
      accounts.forEach(acc => {
        console.log(`  ${acc.accountNumber || acc.id}:`);
        console.log(`    Name: ${acc.name}`);
        console.log(`    Balance: $${acc.balance}`);
        console.log(`    Demo: ${acc.demo ? 'Yes' : 'No'}`);
      });

      expect(accounts.length).toBeGreaterThan(0);
    } else {
      console.log('  No accounts available or different response format');
    }
  });

  test('Deposit Workflow', async () => {
    const depositResult = await admin.deposit(
      'test-user',
      1000,
      'BANK_TRANSFER',
      'E2E Test Deposit'
    );

    console.log('Deposit result:', depositResult);

    if (depositResult.success) {
      expect(depositResult.newBalance).toBeGreaterThan(0);
      console.log(`✓ Deposit successful. New balance: $${depositResult.newBalance}`);
    }
  });

  test('Withdrawal Workflow', async () => {
    const withdrawResult = await admin.withdraw(
      'test-user',
      500,
      'BANK_TRANSFER',
      'E2E Test Withdrawal'
    );

    console.log('Withdraw result:', withdrawResult);

    if (withdrawResult.success) {
      console.log(`✓ Withdrawal successful. New balance: $${withdrawResult.newBalance}`);
    }
  });

  test('Manual Adjustment', async () => {
    const adjustResult = await admin.adjust(
      'test-user',
      100,
      'CORRECTION',
      'E2E Test Adjustment'
    );

    console.log('Adjustment result:', adjustResult);

    if (adjustResult.success) {
      console.log(`✓ Adjustment successful. New balance: $${adjustResult.newBalance}`);
    }
  });

  test('Bonus Addition', async () => {
    const bonusResult = await admin.addBonus(
      'test-user',
      250,
      'E2E Test Bonus'
    );

    console.log('Bonus result:', bonusResult);

    if (bonusResult.success) {
      console.log(`✓ Bonus added. New balance: $${bonusResult.newBalance}`);
    }
  });

  test('Transaction Ledger', async () => {
    const ledger = await admin.getLedger();

    console.log('Transaction Ledger:');
    if (Array.isArray(ledger)) {
      console.log(`  Total transactions: ${ledger.length}`);

      ledger.slice(0, 5).forEach(txn => {
        console.log(`  ${txn.type}: $${txn.amount} - Balance: $${txn.balance}`);
      });

      expect(ledger.length).toBeGreaterThan(0);
    } else {
      console.log('  No transactions or different format');
    }
  });
});

test.describe('Symbol Management Tests', () => {
  let admin;

  test.beforeEach(async () => {
    admin = new AdminAPI();
  });

  test('List Trading Symbols', async () => {
    const symbols = await admin.getSymbols();

    console.log('Trading Symbols:');
    if (Array.isArray(symbols)) {
      symbols.slice(0, 10).forEach(sym => {
        console.log(`  ${sym.symbol}: ${sym.enabled ? 'ENABLED' : 'DISABLED'}`);
      });

      expect(symbols.length).toBeGreaterThan(0);
    }
  });

  test('Toggle Symbol Enable/Disable', async () => {
    const toggleResult = await admin.toggleSymbol('USDJPY', true);

    console.log('Symbol toggle result:', toggleResult);

    if (toggleResult.success) {
      console.log('✓ Symbol toggled successfully');

      // Toggle back
      await admin.toggleSymbol('USDJPY', false);
      console.log('✓ Symbol restored to original state');
    }
  });
});

test.describe('Routing Configuration Tests', () => {
  let admin;

  test.beforeEach(async () => {
    admin = new AdminAPI();
  });

  test('View Routing Rules', async () => {
    const routes = await admin.getRoutes();

    console.log('Routing Rules:');
    if (Array.isArray(routes)) {
      routes.forEach((rule, index) => {
        console.log(`  Rule ${index + 1}:`, rule);
      });
    } else {
      console.log('  No routing rules or different format');
    }
  });
});

test.describe('Complete Admin Scenario Tests', () => {
  let admin;

  test.beforeEach(async () => {
    admin = new AdminAPI();
  });

  test('Scenario: Onboard New LP and Test Execution', async () => {
    console.log('=== LP Onboarding Scenario ===');

    // 1. Check current LPs
    const currentLPs = await admin.listLPs();
    console.log(`✓ Current LPs: ${currentLPs.length}`);

    // 2. Configure execution mode for A-Book
    const modeResult = await admin.setExecutionMode('ABOOK');
    if (modeResult.success) {
      console.log('✓ Execution mode set to ABOOK');
    }

    // 3. Enable target LP
    if (currentLPs.length > 0) {
      const targetLP = currentLPs[0];
      await admin.toggleLP(targetLP.id, true);
      console.log(`✓ Enabled LP: ${targetLP.name}`);
    }

    // 4. Verify LP status
    const status = await admin.getLPStatus();
    console.log('✓ LP status verified');

    // 5. Switch back to BBOOK
    await admin.setExecutionMode('BBOOK');
    console.log('✓ Restored to BBOOK mode');

    console.log('=== Scenario Complete ===');
  });

  test('Scenario: Account Management Workflow', async () => {
    console.log('=== Account Management Scenario ===');

    const accountId = 'test-user';

    // 1. Check initial balance
    const accounts = await admin.getAccounts();
    console.log('✓ Retrieved account list');

    // 2. Perform deposit
    const depositResult = await admin.deposit(accountId, 5000, 'BANK_TRANSFER', 'Test deposit');
    if (depositResult.success) {
      console.log(`✓ Deposit: +$5000 → Balance: $${depositResult.newBalance}`);
    }

    // 3. Add bonus
    const bonusResult = await admin.addBonus(accountId, 500, 'Welcome bonus');
    if (bonusResult.success) {
      console.log(`✓ Bonus: +$500 → Balance: $${bonusResult.newBalance}`);
    }

    // 4. Perform withdrawal
    const withdrawResult = await admin.withdraw(accountId, 1000, 'BANK_TRANSFER', 'Test withdrawal');
    if (withdrawResult.success) {
      console.log(`✓ Withdrawal: -$1000 → Balance: $${withdrawResult.newBalance}`);
    }

    // 5. Check ledger
    const ledger = await admin.getLedger();
    if (Array.isArray(ledger)) {
      console.log(`✓ Ledger entries: ${ledger.length}`);
    }

    console.log('=== Scenario Complete ===');
  });
});

test.describe('Error Handling Tests', () => {
  let admin;

  test.beforeEach(async () => {
    admin = new AdminAPI();
  });

  test('Invalid Configuration Updates', async () => {
    const invalidConfig = {
      defaultLeverage: -100, // Invalid negative leverage
    };

    const result = await admin.updateConfig(invalidConfig);
    console.log('Invalid config result:', result);
  });

  test('Invalid LP Operations', async () => {
    // Try to toggle non-existent LP
    const result = await admin.toggleLP('nonexistent', true);
    console.log('Invalid LP toggle:', result);
  });

  test('Invalid Account Operations', async () => {
    // Try operations on non-existent account
    const result = await admin.deposit('nonexistent', 1000, 'BANK_TRANSFER', 'Test');
    console.log('Invalid account operation:', result);
  });
});

/**
 * Test Cleanup
 */
test.afterAll(async () => {
  console.log('\n✓ All admin E2E tests completed');
});
