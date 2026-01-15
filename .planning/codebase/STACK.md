# Technology Stack

**Analysis Date:** 2026-01-15

## Languages

**Primary:**
- Go 1.24.0 - Backend service (`backend/go.mod`)
- TypeScript 5.9.3 - Frontend applications (`clients/desktop/package.json`, `admin/*/package.json`)

**Secondary:**
- JavaScript - Build configuration files (vite.config.ts, tailwind.config.js)

## Runtime

**Environment:**
- Go 1.24.0 - Backend runtime
- Node.js - Frontend runtime (via Bun)
- Browser - React applications

**Package Manager:**
- Bun - Primary package manager (per CLAUDE.md)
- Go modules - Backend dependency management
- Lockfiles: `bun.lock`, `go.sum`

## Frameworks

**Core:**
- Go standard library - Backend HTTP server, WebSocket, JSON
- React 19.2.0 - Desktop client UI framework
- Next.js 16.1.1 - Admin panel framework
- Vite 7.2.4 - Build tool and dev server

**Testing:**
- Vitest 4.0.17 - Frontend unit testing
- @testing-library/react 16.3.1 - Component testing
- No Go tests found (testing gap)

**Build/Dev:**
- TypeScript 5.9.3 - Type checking
- Vite 7.2.4 - Frontend bundling
- ESLint 9.39.1 - Code linting
- Tailwind CSS 4.1.18 - Styling

## Key Dependencies

**Backend (Go):**
- gorilla/websocket v1.5.3 - WebSocket connections (`backend/ws/hub.go`, `backend/binance/client.go`)
- golang-jwt/jwt v5.3.0 - JWT authentication (`backend/auth/token.go`)
- golang.org/x/crypto v0.46.0 - Password hashing (`backend/auth/service.go`)
- google/uuid v1.6.0 - UUID generation

**Frontend (React):**
- lightweight-charts 5.1.0 - Financial charting (`clients/desktop/src/components/TradingChart.tsx`)
- technicalindicators 3.1.0 - Technical analysis (`clients/desktop/src/indicators/core/IndicatorEngine.ts`)
- lucide-react 0.562.0 - Icon library
- Tailwind CSS 4.1.18 - Styling

## Configuration

**Environment:**
- Hardcoded configuration in `backend/cmd/server/main.go` (security concern - no .env files)
- LP configuration: `backend/data/lp_config.json`
- Frontend: Hardcoded localhost URLs (configuration gap)

**Build:**
- `go.mod` / `go.sum` - Go dependencies
- `package.json` / `bun.lock` - Node dependencies
- `tsconfig.json` - TypeScript compiler config
- `vite.config.ts` - Vite bundler config
- `vitest.config.ts` - Test runner config
- `tailwind.config.js` - Tailwind CSS config

## Platform Requirements

**Development:**
- Any platform (macOS, Linux, Windows)
- Go 1.24.0+ installed
- Bun package manager

**Production:**
- Go binary compiled for target platform
- Static frontend assets served via HTTP

---

*Stack analysis: 2026-01-15*
*Update after major dependency changes*
