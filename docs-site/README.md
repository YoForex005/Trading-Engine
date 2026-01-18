# Trading Platform Documentation

This is the comprehensive documentation website for Trading Platform, built with [Docusaurus 3](https://docusaurus.io/).

## Quick Start

### Installation

```bash
npm install
```

### Local Development

```bash
npm start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

### Build

```bash
npm run build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

### Deployment

#### Vercel

```bash
npm install -g vercel
vercel
```

#### Netlify

```bash
npm install -g netlify-cli
netlify deploy
```

## Documentation Structure

```
docs/
├── getting-started/      # Getting started guides
├── trading-guide/        # Trading tutorials and guides
├── api/                  # API documentation
│   ├── rest/            # REST API reference
│   ├── websocket/       # WebSocket API reference
│   ├── fix44/           # FIX 4.4 protocol docs
│   └── examples/        # Code examples
├── admin/               # Admin panel documentation
├── integrations/        # Third-party integrations
├── reference/           # Technical reference
├── troubleshooting/     # Common issues and FAQs
└── legal/               # Legal documents
```

## Features

- Full REST, WebSocket, and FIX 4.4 API documentation
- Interactive code examples in multiple languages
- Dark mode support
- Multi-language support (i18n)
- Fast search with Algolia
- Versioned documentation
- Mobile responsive
- OpenAPI integration

## Technologies

- **Docusaurus 3**: Modern static site generator
- **TypeScript**: Type-safe configuration
- **MDX**: Interactive documentation
- **OpenAPI**: Auto-generated API docs
- **Algolia**: Fast search functionality
- **Vercel/Netlify**: Deployment platforms

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

Copyright © 2026 Your Trading Platform, Inc. All rights reserved.

## Support

- Email: docs@yourtradingplatform.com
- Discord: https://discord.gg/yourtradingplatform
- Website: https://yourtradingplatform.com
