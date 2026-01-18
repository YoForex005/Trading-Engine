# FIX Gateway - YoFx Integration

Professional-grade FIX 4.4 protocol implementation for YoFx liquidity provider integration.

## 📁 Project Structure

This project follows industry best practices for scalable financial trading systems:

```
fix/
├── cmd/                    # Application entry points
│   └── tests/             # Test executables and validation scripts
│       ├── test_all_features.go        # Comprehensive feature testing
│       ├── test_connection.go          # Basic connectivity test
│       ├── test_order_placement.go     # Order placement validation
│       ├── test_securities_list.go     # Symbol discovery test
│       └── test_yofx2_marketdata.go    # Market data feed test
│
├── config/                # Configuration files
│   ├── sessions.json      # FIX session configurations (both YOFX1 & YOFX2)
│   └── yofx1_session.cfg  # Legacy INI-style config
│
├── docs/                  # Documentation
│   ├── README.md          # This file (main documentation)
│   ├── deploy_vps.md      # VPS deployment guide
│   └── test_summary.md    # Test results and findings
│
├── internal/              # Private application code
│   ├── gateway/           # FIX gateway core implementation
│   ├── session/           # Session management logic
│   └── message/           # FIX message handlers
│
├── pkg/                   # Public, reusable packages
│   ├── fixclient/         # FIX client library
│   └── types/             # Shared types and interfaces
│
├── scripts/               # Utility scripts
│   └── ssh_tunnel.sh      # SSH tunnel for port forwarding
│
├── examples/              # Usage examples and sample code
│
├── backup/                # Backup directory for old files
│
└── gateway.go             # Main gateway implementation (to be refactored)
```

## 🎯 Architecture Overview

### Design Principles

This project follows these architectural patterns:

1. **Separation of Concerns**: Clear separation between configuration, business logic, and tests
2. **Modular Design**: Each component (gateway, session, message) is self-contained
3. **Scalability**: Structure supports adding multiple LPs, protocols, and features
4. **Maintainability**: Intuitive organization for both developers and business stakeholders

### Directory Purposes

#### `cmd/tests/`
**Purpose**: Executable test programs
**Why**: Isolated test executables prevent mixing test code with production code. Each test can be run independently.

**Files**:
- `test_all_features.go` - Comprehensive YOFX1 feature validation
- `test_yofx2_marketdata.go` - YOFX2 market data feed testing
- `test_connection.go` - Basic connectivity verification
- `test_order_placement.go` - Order execution testing ⚠️
- `test_securities_list.go` - Symbol/instrument discovery

#### `config/`
**Purpose**: Configuration files for all environments
**Why**: Centralized configuration management, easy to version control and deploy

**Files**:
- `sessions.json` - Primary configuration (both sessions)
- `yofx1_session.cfg` - Legacy format (for compatibility)

#### `docs/`
**Purpose**: All documentation and guides
**Why**: Keeps documentation separate from code, easy to find and update

**Files**:
- `README.md` - Main project documentation
- `deploy_vps.md` - Deployment instructions
- `test_summary.md` - Test results and API capabilities

#### `internal/`
**Purpose**: Private application code (not importable by external packages)
**Why**: Enforces encapsulation, prevents external dependencies on internal implementation

**Planned Structure**:
- `gateway/` - Core FIX gateway logic
- `session/` - Session lifecycle management
- `message/` - FIX message parsing and construction

#### `pkg/`
**Purpose**: Public, reusable packages
**Why**: Can be imported by other projects, promotes code reuse

**Planned Structure**:
- `fixclient/` - Reusable FIX client library
- `types/` - Shared types, interfaces, constants

#### `scripts/`
**Purpose**: Operational scripts and utilities
**Why**: Separate automation from application code

**Files**:
- `ssh_tunnel.sh` - Create SSH tunnel for blocked ports

#### `examples/`
**Purpose**: Sample code and usage patterns
**Why**: Helps developers understand how to use the library

---

## 🔧 Sessions Overview

### YOFX1 - Trading Operations
**Purpose**: Order placement, execution, and trading operations

**Configuration**:
- **SenderCompID**: YOFX1
- **TargetCompID**: YOFX
- **Account**: 50153
- **Server**: 23.106.238.138:12336

**Capabilities**:
- Place/modify/cancel orders
- Request positions
- Query account information

### YOFX2 - Market Data Feed
**Purpose**: Real-time market data, quotes, and prices

**Configuration**:
- **SenderCompID**: YOFX2
- **TargetCompID**: YOFX
- **Account**: 50153
- **Server**: 23.106.238.138:12336

**Capabilities**:
- Subscribe to real-time quotes
- Receive market data snapshots
- Monitor price updates

---

## 🚀 Quick Start

### Run Tests

```bash
# Test YOFX1 (Trading)
cd cmd/tests
go run test_all_features.go

# Test YOFX2 (Market Data)
go run test_yofx2_marketdata.go

# Test Basic Connectivity
go run test_connection.go
```

### Network Troubleshooting

If experiencing connection timeouts:

```bash
# 1. Test port accessibility
nc -zv 23.106.238.138 12336

# 2. Use SSH tunnel (if port blocked)
cd scripts
./ssh_tunnel.sh user@your-vps-ip

# 3. Then update tests to use localhost:12336
```

---

## 📊 Test Results Summary

See `docs/test_summary.md` for detailed test results.

**Quick Status**:
- ✅ YOFX1 connection working
- ✅ YOFX2 connection working
- ✅ Position requests working
- ⚠️ Market data subscriptions accepted (no quotes yet)
- ❌ Crypto symbols not available (BTCUSD, ETHUSD)

---

## 🔄 Migration Plan

### Current State
All code is in the root `fix/` directory - difficult to maintain and scale.

### Target State (Refactoring Roadmap)

1. **Phase 1: Structure** ✅ DONE
   - Create organized folder structure
   - Move files to appropriate directories

2. **Phase 2: Modularization** (NEXT)
   - Extract `gateway.go` into `internal/gateway/`
   - Create session management in `internal/session/`
   - Build message handlers in `internal/message/`

3. **Phase 3: Library Extraction**
   - Create reusable `pkg/fixclient/`
   - Define types in `pkg/types/`
   - Add examples in `examples/`

4. **Phase 4: Production Ready**
   - Add comprehensive error handling
   - Implement reconnection logic
   - Add sequence number persistence
   - Create deployment automation

---

## 🛠️ Development Guidelines

### Adding New Tests
1. Create test file in `cmd/tests/`
2. Use `// +build ignore` at the top
3. Follow naming: `test_<feature>.go`
4. Document findings in `docs/test_summary.md`

### Adding New Sessions
1. Add configuration to `config/sessions.json`
2. Update `docs/README.md` with session details
3. Create corresponding test in `cmd/tests/`

### Modifying Gateway
1. Code goes in `internal/gateway/`
2. Keep public interfaces in `pkg/`
3. Update tests to verify changes

---

## 📚 References

Based on industry best practices from:
- [Organize Like a Pro: Go Project Folder Structures](https://medium.com/@smart_byte_labs/organize-like-a-pro-a-simple-guide-to-go-project-folder-structures-e85e9c1769c2)
- [11 Tips for Structuring Go Projects - Alex Edwards](https://www.alexedwards.net/blog/11-tips-for-structuring-your-go-projects)
- [Optimal Project Layout for Large-Scale Go Applications](https://leapcell.io/blog/optimal-project-layout-for-large-scale-go-applications)
- [Best Practices for Go Project Structure](https://medium.com/@nandoseptian/best-practices-for-go-project-structure-and-code-organization-486898990d0a)
- [Go Project Structure Best Practices - TutorialEdge](https://tutorialedge.net/golang/go-project-structure-best-practices/)

---

## 📝 License

Internal use only - YoFx FIX integration for trading engine backend.

---

## 🤝 Contributing

When contributing:
1. Follow the folder structure guidelines
2. Add tests for new features in `cmd/tests/`
3. Update documentation in `docs/`
4. Keep configuration in `config/`
5. Ensure backward compatibility

---

**Last Updated**: 2026-01-18
**Status**: Structure reorganized, ready for Phase 2 refactoring
