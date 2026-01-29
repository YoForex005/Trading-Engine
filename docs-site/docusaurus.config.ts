import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  title: 'Trading Platform Documentation',
  tagline: 'Professional Trading Platform - Comprehensive Documentation',
  favicon: 'img/favicon.ico',

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },

  // Set the production url of your site here
  url: 'https://docs.yourtradingplatform.com',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'your-trading-platform', // Usually your GitHub org/user name.
  projectName: 'trading-platform-docs', // Usually your repo name.

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'zh', 'es', 'fr', 'de', 'ja'],
    localeConfigs: {
      en: {
        label: 'English',
        direction: 'ltr',
        htmlLang: 'en-US',
      },
      zh: {
        label: '简体中文',
        direction: 'ltr',
        htmlLang: 'zh-CN',
      },
      es: {
        label: 'Español',
        direction: 'ltr',
        htmlLang: 'es-ES',
      },
    },
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          routeBasePath: 'docs',
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/your-trading-platform/trading-platform-docs/tree/main/',
          showLastUpdateAuthor: true,
          showLastUpdateTime: true,
          remarkPlugins: [],
          rehypePlugins: [],
          // Versioning
          versions: {
            current: {
              label: 'v2.0 (Current)',
              path: '',
            },
          },
          lastVersion: 'current',
          onlyIncludeVersions: ['current'],
        },
        blog: {
          showReadingTime: true,
          feedOptions: {
            type: ['rss', 'atom'],
            xslt: true,
          },
          blogTitle: 'Platform Updates & News',
          blogDescription: 'Stay updated with the latest features, updates, and trading insights',
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/your-trading-platform/trading-platform-docs/tree/main/',
          // Useful options to enforce blogging best practices
          onInlineTags: 'warn',
          onInlineAuthors: 'warn',
          onUntruncatedBlogPosts: 'warn',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
        googleAnalytics: {
          trackingID: 'G-XXXXXXXXXX', // Replace with your GA tracking ID
          anonymizeIP: true,
        },
        gtag: {
          trackingID: 'G-XXXXXXXXXX', // Replace with your GA tracking ID
          anonymizeIP: true,
        },
      } satisfies Preset.Options,
    ],
  ],

  plugins: [
    [
      'docusaurus-plugin-openapi-docs',
      {
        id: 'api',
        docsPluginId: 'classic',
        config: {
          tradingApi: {
            specPath: '../backend/api/openapi.yaml', // Path to your OpenAPI spec
            outputDir: 'docs/api',
            sidebarOptions: {
              groupPathsBy: 'tag',
              categoryLinkSource: 'tag',
            },
          },
        },
      },
    ],
    '@docusaurus/plugin-ideal-image',
    'docusaurus-plugin-image-zoom',
  ],

  themes: ['docusaurus-theme-openapi-docs'],

  themeConfig: {
    // Replace with your project's social card
    image: 'img/social-card.jpg',
    colorMode: {
      defaultMode: 'light',
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },
    docs: {
      sidebar: {
        hideable: true,
        autoCollapseCategories: true,
      },
    },
    // Algolia search configuration
    algolia: {
      appId: 'YOUR_APP_ID', // Replace with your Algolia app ID
      apiKey: 'YOUR_SEARCH_API_KEY', // Replace with your Algolia search API key
      indexName: 'trading-platform',
      contextualSearch: true,
      searchPagePath: 'search',
      insights: true,
    },
    navbar: {
      title: 'Trading Platform',
      hideOnScroll: true,
      logo: {
        alt: 'Trading Platform Logo',
        src: 'img/logo.svg',
        srcDark: 'img/logo-dark.svg',
        height: 32,
        width: 32,
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'gettingStartedSidebar',
          position: 'left',
          label: 'Getting Started',
        },
        {
          type: 'docSidebar',
          sidebarId: 'tradingGuideSidebar',
          position: 'left',
          label: 'Trading Guide',
        },
        {
          type: 'docSidebar',
          sidebarId: 'apiSidebar',
          position: 'left',
          label: 'API',
        },
        {
          type: 'docSidebar',
          sidebarId: 'adminSidebar',
          position: 'left',
          label: 'Admin',
        },
        {
          type: 'docSidebar',
          sidebarId: 'integrationsSidebar',
          position: 'left',
          label: 'Integrations',
        },
        {
          type: 'dropdown',
          label: 'More',
          position: 'left',
          items: [
            {
              type: 'docSidebar',
              sidebarId: 'referenceSidebar',
              label: 'Reference',
            },
            {
              type: 'docSidebar',
              sidebarId: 'troubleshootingSidebar',
              label: 'Troubleshooting',
            },
            {
              type: 'docSidebar',
              sidebarId: 'legalSidebar',
              label: 'Legal',
            },
          ],
        },
        {to: '/blog', label: 'Updates', position: 'left'},
        {
          type: 'docsVersionDropdown',
          position: 'right',
          dropdownActiveClassDisabled: true,
        },
        {
          type: 'localeDropdown',
          position: 'right',
        },
        {
          href: 'https://github.com/your-trading-platform/trading-platform',
          position: 'right',
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
        },
        {
          href: 'https://status.yourtradingplatform.com',
          position: 'right',
          label: 'Status',
        },
      ],
    },
    footer: {
      style: 'dark',
      logo: {
        alt: 'Trading Platform Logo',
        src: 'img/logo.svg',
        height: 36,
      },
      links: [
        {
          title: 'Documentation',
          items: [
            {
              label: 'Getting Started',
              to: '/docs/getting-started/introduction',
            },
            {
              label: 'API Reference',
              to: '/docs/api/overview',
            },
            {
              label: 'Trading Guide',
              to: '/docs/trading-guide/overview',
            },
            {
              label: 'Admin Guide',
              to: '/docs/admin/overview',
            },
          ],
        },
        {
          title: 'Resources',
          items: [
            {
              label: 'Integrations',
              to: '/docs/integrations/overview',
            },
            {
              label: 'SDKs & Libraries',
              to: '/docs/api/sdks',
            },
            {
              label: 'Troubleshooting',
              to: '/docs/troubleshooting/faq',
            },
            {
              label: 'Platform Status',
              href: 'https://status.yourtradingplatform.com',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Discord',
              href: 'https://discord.gg/yourtradingplatform',
            },
            {
              label: 'Twitter',
              href: 'https://twitter.com/yourtradingplatform',
            },
            {
              label: 'Stack Overflow',
              href: 'https://stackoverflow.com/questions/tagged/your-trading-platform',
            },
            {
              label: 'Community Forum',
              href: 'https://community.yourtradingplatform.com',
            },
          ],
        },
        {
          title: 'Legal & Support',
          items: [
            {
              label: 'Terms of Service',
              to: '/docs/legal/terms-of-service',
            },
            {
              label: 'Privacy Policy',
              to: '/docs/legal/privacy-policy',
            },
            {
              label: 'Risk Disclosure',
              to: '/docs/legal/risk-disclosure',
            },
            {
              label: 'Support Center',
              href: 'https://support.yourtradingplatform.com',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Your Trading Platform, Inc. All rights reserved. Trading involves risk. See <a href="/docs/legal/risk-disclosure">Risk Disclosure</a>.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'json', 'python', 'javascript', 'typescript', 'go', 'java', 'csharp', 'php', 'ruby'],
    },
    zoom: {
      selector: '.markdown :not(em) > img',
      background: {
        light: 'rgb(255, 255, 255)',
        dark: 'rgb(50, 50, 50)',
      },
      config: {
        // Zoom configuration options
      },
    },
    announcementBar: {
      id: 'announcement-bar',
      content:
        '🚀 New: FIX 4.4 API now available! <a target="_blank" rel="noopener noreferrer" href="/docs/api/fix44/overview">Learn more</a>',
      backgroundColor: '#4CAF50',
      textColor: '#FFFFFF',
      isCloseable: true,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
