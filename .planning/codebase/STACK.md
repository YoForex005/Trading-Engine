# Technology Stack

**Analysis Date:** 2026-01-18

## Languages

**Primary:**
- Go 1.24.0 - Backend trading engine, FIX gateway, WebSocket server, order management
- TypeScript 5.9.3 - Desktop client frontend

**Secondary:**
- JavaScript (ES modules) - Build tooling, configuration files

## Runtime

**Environment:**
- Go 1.25.5 (darwin/arm64) - Backend runtime
- Node.js v24.11.1 - Frontend build and development

**Package Manager:**
- Go modules (go.mod/go.sum) - Backend dependency management
- npm with package-lock.json - Frontend dependency management
- Lockfile: Present for both ecosystems

## Frameworks

**Core:**
- Vite 7.2.4 - Frontend build tool and dev server
- React 19.2.0 - Desktop client UI framework
- Gorilla WebSocket 1.5.3 - Real-time market data streaming

**Testing:**
- Not detected - No test framework currently configured

**Build/Dev:**
- Vite 7.2.4 - Frontend bundler with HMR
- TypeScript compiler (tsc) - Type checking
- ESLint 9.39.1 - Frontend linting
- Tailwind CSS 4.1.18 - UI styling

## Key Dependencies

**Critical:**
- `github.com/gorilla/websocket` v1.5.3 - WebSocket connections for real-time price feeds and client communication
- `github.com/golang-jwt/jwt/v5` v5.3.0 - Authentication token generation and validation
- `golang.org/x/crypto` v0.46.0 - bcrypt password hashing for user authentication
- `github.com/google/uuid` v1.6.0 - Order ID and account ID generation
- `react` v19.2.0 - Desktop client UI components
- `lightweight-charts` v5.1.0 - Financial charting for price visualization

**Infrastructure:**
- `@vitejs/plugin-react` v5.1.1 - React JSX transformation
- `tailwindcss` v4.1.18 - Utility-first CSS framework
- `lucide-react` v0.562.0 - Icon library for UI components
- `typescript-eslint` v8.46.4 - TypeScript linting rules

## Configuration

**Environment:**
- OANDA API key and account ID hardcoded in `backend/cmd/server/main.go` (const OANDA_API_KEY, OANDA_ACCOUNT_ID)
- FIX session credentials stored in `backend/fix/sessions.json`
- LP configuration in `backend/data/lp_config.json`
- No .env files in active use (found in Trading-Engine subdirectory but not current project)

**Build:**
- `backend/go.mod` - Go module configuration
- `clients/desktop/vite.config.ts` - Vite build configuration
- `clients/desktop/tsconfig.json` - TypeScript project references
- `clients/desktop/tsconfig.app.json` - Application TypeScript settings
- `clients/desktop/eslint.config.js` - Linting rules
- `clients/desktop/tailwind.config.js` - CSS framework configuration
- `clients/desktop/postcss.config.js` - CSS processing

## Platform Requirements

**Development:**
- Go v1.22+ (currently using v1.24.0)
- Node.js v18+ (currently using v24.11.1)
- npm (bundled with Node.js)
- macOS (darwin/arm64) or compatible platform
- Git LFS for large tick data files

**Production:**
- Deployment target: Not specified (no Docker, Makefile, or deployment configs detected)
- Database: PostgreSQL with TimescaleDB extension recommended (schema defined in `backend/database/schema.sql`)
- WebSocket-capable reverse proxy (nginx/Caddy) for production WS connections
- TLS certificates for secure FIX protocol connections (optional, SSL flag in sessions config)

---

*Stack analysis: 2026-01-18*
