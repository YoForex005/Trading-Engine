import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  // Getting Started Sidebar
  gettingStartedSidebar: [
    {
      type: 'category',
      label: 'Getting Started',
      collapsed: false,
      items: [
        'getting-started/introduction',
        'getting-started/quick-start',
        'getting-started/installation',
        'getting-started/account-setup',
        'getting-started/first-trade',
        'getting-started/platform-overview',
        'getting-started/key-concepts',
      ],
    },
  ],

  // Trading Guide Sidebar
  tradingGuideSidebar: [
    {
      type: 'category',
      label: 'Trading Guide',
      collapsed: false,
      items: [
        'trading-guide/overview',
        {
          type: 'category',
          label: 'Orders',
          items: [
            'trading-guide/orders/placing-orders',
            'trading-guide/orders/order-types',
            'trading-guide/orders/market-orders',
            'trading-guide/orders/limit-orders',
            'trading-guide/orders/stop-orders',
            'trading-guide/orders/conditional-orders',
            'trading-guide/orders/order-execution',
          ],
        },
        {
          type: 'category',
          label: 'Position Management',
          items: [
            'trading-guide/positions/overview',
            'trading-guide/positions/opening-positions',
            'trading-guide/positions/closing-positions',
            'trading-guide/positions/modifying-positions',
            'trading-guide/positions/hedging',
          ],
        },
        {
          type: 'category',
          label: 'Risk Management',
          items: [
            'trading-guide/risk/overview',
            'trading-guide/risk/stop-loss',
            'trading-guide/risk/take-profit',
            'trading-guide/risk/position-sizing',
            'trading-guide/risk/risk-reward-ratio',
            'trading-guide/risk/risk-calculator',
          ],
        },
        {
          type: 'category',
          label: 'Margin & Leverage',
          items: [
            'trading-guide/margin/margin-explained',
            'trading-guide/margin/margin-requirements',
            'trading-guide/margin/margin-calls',
            'trading-guide/margin/leverage-explained',
            'trading-guide/margin/leverage-calculator',
          ],
        },
        {
          type: 'category',
          label: 'Trading Strategies',
          items: [
            'trading-guide/strategies/day-trading',
            'trading-guide/strategies/swing-trading',
            'trading-guide/strategies/scalping',
            'trading-guide/strategies/algorithmic-trading',
            'trading-guide/strategies/copy-trading',
          ],
        },
      ],
    },
  ],

  // API Sidebar
  apiSidebar: [
    {
      type: 'category',
      label: 'API Documentation',
      collapsed: false,
      items: [
        'api/overview',
        'api/authentication',
        'api/rate-limits',
        'api/errors',
        'api/sdks',
        {
          type: 'category',
          label: 'REST API',
          items: [
            'api/rest/overview',
            'api/rest/accounts',
            'api/rest/orders',
            'api/rest/positions',
            'api/rest/market-data',
            'api/rest/history',
            'api/rest/reporting',
          ],
        },
        {
          type: 'category',
          label: 'WebSocket API',
          items: [
            'api/websocket/overview',
            'api/websocket/connection',
            'api/websocket/authentication',
            'api/websocket/market-data',
            'api/websocket/order-updates',
            'api/websocket/account-updates',
          ],
        },
        {
          type: 'category',
          label: 'FIX 4.4 API',
          items: [
            'api/fix44/overview',
            'api/fix44/connection',
            'api/fix44/authentication',
            'api/fix44/messages',
            'api/fix44/order-entry',
            'api/fix44/market-data',
            'api/fix44/session-management',
          ],
        },
        {
          type: 'category',
          label: 'Code Examples',
          items: [
            'api/examples/python',
            'api/examples/javascript',
            'api/examples/go',
            'api/examples/java',
            'api/examples/csharp',
          ],
        },
      ],
    },
  ],

  // Admin Guide Sidebar
  adminSidebar: [
    {
      type: 'category',
      label: 'Admin Guide',
      collapsed: false,
      items: [
        'admin/overview',
        'admin/getting-started',
        {
          type: 'category',
          label: 'User Management',
          items: [
            'admin/users/overview',
            'admin/users/creating-users',
            'admin/users/user-permissions',
            'admin/users/kyc-verification',
            'admin/users/account-status',
          ],
        },
        {
          type: 'category',
          label: 'LP Management',
          items: [
            'admin/lp/overview',
            'admin/lp/adding-lps',
            'admin/lp/lp-configuration',
            'admin/lp/monitoring',
          ],
        },
        {
          type: 'category',
          label: 'Routing Configuration',
          items: [
            'admin/routing/overview',
            'admin/routing/routing-rules',
            'admin/routing/symbol-routing',
            'admin/routing/failover',
          ],
        },
        {
          type: 'category',
          label: 'System Monitoring',
          items: [
            'admin/monitoring/dashboard',
            'admin/monitoring/alerts',
            'admin/monitoring/logs',
            'admin/monitoring/metrics',
          ],
        },
        {
          type: 'category',
          label: 'Backup & Recovery',
          items: [
            'admin/backup/overview',
            'admin/backup/backup-strategy',
            'admin/backup/restore-procedures',
            'admin/backup/disaster-recovery',
          ],
        },
      ],
    },
  ],

  // Integrations Sidebar
  integrationsSidebar: [
    {
      type: 'category',
      label: 'Integrations',
      collapsed: false,
      items: [
        'integrations/overview',
        {
          type: 'category',
          label: 'MetaTrader',
          items: [
            'integrations/metatrader/overview',
            'integrations/metatrader/mt4-setup',
            'integrations/metatrader/mt5-setup',
            'integrations/metatrader/bridge-configuration',
          ],
        },
        {
          type: 'category',
          label: 'TradingView',
          items: [
            'integrations/tradingview/overview',
            'integrations/tradingview/setup',
            'integrations/tradingview/webhooks',
            'integrations/tradingview/strategies',
          ],
        },
        {
          type: 'category',
          label: 'Algorithmic Trading',
          items: [
            'integrations/algo/overview',
            'integrations/algo/setup',
            'integrations/algo/backtesting',
            'integrations/algo/deployment',
          ],
        },
        {
          type: 'category',
          label: 'Webhooks',
          items: [
            'integrations/webhooks/overview',
            'integrations/webhooks/setup',
            'integrations/webhooks/events',
            'integrations/webhooks/security',
          ],
        },
        {
          type: 'category',
          label: 'Payment Gateways',
          items: [
            'integrations/payments/overview',
            'integrations/payments/stripe',
            'integrations/payments/paypal',
            'integrations/payments/crypto-payments',
          ],
        },
      ],
    },
  ],

  // Reference Sidebar
  referenceSidebar: [
    {
      type: 'category',
      label: 'Reference',
      collapsed: false,
      items: [
        'reference/overview',
        {
          type: 'category',
          label: 'Symbols & Instruments',
          items: [
            'reference/symbols/forex',
            'reference/symbols/cryptocurrencies',
            'reference/symbols/commodities',
            'reference/symbols/indices',
            'reference/symbols/stocks',
          ],
        },
        {
          type: 'category',
          label: 'Trading Hours',
          items: [
            'reference/hours/overview',
            'reference/hours/forex-hours',
            'reference/hours/crypto-hours',
            'reference/hours/commodity-hours',
          ],
        },
        {
          type: 'category',
          label: 'Fees & Commissions',
          items: [
            'reference/fees/overview',
            'reference/fees/spreads',
            'reference/fees/commissions',
            'reference/fees/swap-rates',
          ],
        },
        {
          type: 'category',
          label: 'Margin Requirements',
          items: [
            'reference/margin/overview',
            'reference/margin/by-instrument',
            'reference/margin/calculation',
          ],
        },
        {
          type: 'category',
          label: 'Contract Specifications',
          items: [
            'reference/contracts/overview',
            'reference/contracts/forex-specs',
            'reference/contracts/crypto-specs',
            'reference/contracts/commodity-specs',
          ],
        },
      ],
    },
  ],

  // Troubleshooting Sidebar
  troubleshootingSidebar: [
    {
      type: 'category',
      label: 'Troubleshooting',
      collapsed: false,
      items: [
        'troubleshooting/overview',
        {
          type: 'category',
          label: 'Common Issues',
          items: [
            'troubleshooting/common/login-issues',
            'troubleshooting/common/order-execution',
            'troubleshooting/common/connection-problems',
            'troubleshooting/common/api-errors',
          ],
        },
        {
          type: 'category',
          label: 'Error Codes',
          items: [
            'troubleshooting/errors/overview',
            'troubleshooting/errors/http-errors',
            'troubleshooting/errors/websocket-errors',
            'troubleshooting/errors/fix-errors',
          ],
        },
        'troubleshooting/faq',
        'troubleshooting/support',
      ],
    },
  ],

  // Legal Sidebar
  legalSidebar: [
    {
      type: 'category',
      label: 'Legal',
      collapsed: false,
      items: [
        'legal/terms-of-service',
        'legal/privacy-policy',
        'legal/risk-disclosure',
        'legal/cookie-policy',
        'legal/aml-policy',
        'legal/data-protection',
        'legal/complaints-procedure',
      ],
    },
  ],
};

export default sidebars;
