# Project Structure Guide

## Why This Structure?

This folder organization follows **industry best practices** for scalable financial trading systems, based on:

1. **Go Standard Project Layout** - Used by major Go projects
2. **Domain-Driven Design** - Organized around business capabilities
3. **Clean Architecture** - Separation of concerns and dependencies
4. **Financial System Patterns** - Common patterns in trading platforms

---

## Directory Breakdown

### 📂 `cmd/` - Application Entry Points

**Purpose**: Main applications and executables

**Why Separate?**
- Each program has its own directory
- Easy to build multiple executables from one project
- Clear distinction between different tools

**Current Usage**:
```
cmd/
└── tests/          # Test executables (not unit tests)
```

**Best Practices**:
- One executable per subdirectory
- Minimal code - delegate to `internal/` or `pkg/`
- Only contains `main.go` or startup logic

**Example Future Structure**:
```
cmd/
├── gateway/        # Main FIX gateway server
├── trader/         # Trading bot application
└── tests/          # Testing utilities
```

---

### 📂 `internal/` - Private Application Code

**Purpose**: Private code that cannot be imported by external projects

**Why Use Internal?**
- Go enforces `internal/` as non-importable
- Prevents external dependencies on implementation details
- Freedom to refactor without breaking external users
- Encapsulation at the language level

**Current Structure**:
```
internal/
├── gateway/        # FIX gateway core logic
├── session/        # Session lifecycle management
└── message/        # FIX message handlers
```

**What Goes Here?**:
- Core business logic
- Session management
- Message parsing/construction
- Internal utilities
- Domain models

**Best Practices**:
- Organize by domain/feature, not by type
- Keep packages focused and cohesive
- Avoid circular dependencies

---

### 📂 `pkg/` - Public Library Code

**Purpose**: Code that can be imported by external projects

**Why Separate from Internal?**
- Clearly signals "this is reusable"
- Can be imported by other teams/projects
- Requires stable API contracts
- Promotes code reuse

**Planned Structure**:
```
pkg/
├── fixclient/      # Reusable FIX client library
│   ├── client.go   # Main client interface
│   ├── session.go  # Session management
│   └── message.go  # Message utilities
└── types/          # Shared types and constants
    ├── orders.go   # Order types
    ├── quotes.go   # Quote types
    └── common.go   # Common types
```

**What Goes Here?**:
- Reusable libraries
- Common types and interfaces
- Utility functions
- SDK-like components

**Best Practices**:
- API stability is important
- Good documentation required
- Version carefully (breaking changes)

---

### 📂 `config/` - Configuration Files

**Purpose**: All configuration files for different environments

**Why Centralize?**
- Easy to find all configs
- Simplifies deployment
- Clear separation from code
- Version control friendly

**Current Files**:
```
config/
├── sessions.json           # FIX session configs
└── yofx1_session.cfg      # Legacy config format
```

**Future Structure**:
```
config/
├── sessions/
│   ├── yofx1.json         # YOFX1 trading session
│   ├── yofx2.json         # YOFX2 market data
│   └── lp2.json           # Additional LP
├── environments/
│   ├── dev.json
│   ├── staging.json
│   └── production.json
└── schema.json            # Config validation schema
```

**Best Practices**:
- Never commit secrets (use env vars or secret managers)
- Provide example configs (`.example.json`)
- Validate configs on startup
- Support multiple formats (JSON, YAML, TOML)

---

### 📂 `docs/` - Documentation

**Purpose**: All project documentation

**Why Separate?**
- Easy to find and browse
- Can be hosted separately (wiki, docs site)
- Doesn't clutter code directories
- Better organization for different doc types

**Current Files**:
```
docs/
├── README.md              # Main documentation
├── deploy_vps.md          # Deployment guide
└── test_summary.md        # Test results
```

**Future Structure**:
```
docs/
├── README.md              # Project overview
├── architecture/
│   ├── overview.md        # System architecture
│   ├── fix-protocol.md    # FIX implementation details
│   └── data-flow.md       # How data flows
├── api/
│   ├── gateway-api.md     # Gateway API reference
│   └── fix-messages.md    # Supported FIX messages
├── guides/
│   ├── getting-started.md
│   ├── deployment.md
│   └── troubleshooting.md
└── decisions/             # Architecture Decision Records (ADR)
    ├── 001-folder-structure.md
    └── 002-session-management.md
```

**Best Practices**:
- Keep docs close to code (in repo)
- Use markdown for portability
- Include diagrams (mermaid, plantuml)
- Document decisions (ADRs)

---

### 📂 `scripts/` - Utility Scripts

**Purpose**: Operational scripts and automation

**Why Separate?**
- Clear distinction from application code
- Easy to find operational tools
- Can be used in CI/CD
- Shell/Python/etc scripts in one place

**Current Files**:
```
scripts/
└── ssh_tunnel.sh          # SSH tunnel setup
```

**Future Structure**:
```
scripts/
├── dev/
│   ├── setup.sh           # Development setup
│   └── mock-server.sh     # Start mock FIX server
├── deploy/
│   ├── deploy-vps.sh      # VPS deployment
│   └── rollback.sh        # Rollback deployment
├── ops/
│   ├── health-check.sh    # System health check
│   └── logs-tail.sh       # View logs
└── tunnel/
    └── ssh_tunnel.sh      # SSH tunnel
```

**Best Practices**:
- Make scripts executable (`chmod +x`)
- Add shebang (`#!/bin/bash`)
- Include help text (`--help`)
- Make idempotent when possible

---

### 📂 `examples/` - Usage Examples

**Purpose**: Sample code showing how to use the system

**Why Include?**
- Helps developers understand usage
- Serves as documentation
- Can be used in tutorials
- Validates library design

**Future Structure**:
```
examples/
├── basic-connection/
│   └── main.go            # Simple connection example
├── market-data/
│   └── main.go            # Subscribe to market data
├── order-placement/
│   └── main.go            # Place and manage orders
└── custom-strategy/
    └── main.go            # Complete trading strategy
```

**Best Practices**:
- Keep examples simple and focused
- Include comments explaining each step
- Make examples runnable
- Cover common use cases

---

### 📂 `backup/` - Backup Files

**Purpose**: Temporary storage for old files during reorganization

**When to Use**:
- During major refactoring
- Keeping old versions temporarily
- Migration period

**Important**:
- Should be temporary
- Add to `.gitignore` if needed
- Clean up after migration complete

---

## File Placement Guidelines

### "Where should this file go?"

Use this decision tree:

```
Is it an executable program?
├─ Yes → cmd/<program-name>/
└─ No
   │
   Is it reusable by other projects?
   ├─ Yes → pkg/<package-name>/
   └─ No
      │
      Is it application-specific logic?
      ├─ Yes → internal/<domain>/
      └─ No
         │
         Is it configuration?
         ├─ Yes → config/
         └─ No
            │
            Is it documentation?
            ├─ Yes → docs/
            └─ No
               │
               Is it a script/tool?
               ├─ Yes → scripts/
               └─ No → examples/ or root
```

---

## Common Questions

### Q: Should tests go in `cmd/tests/`?

**A**: Depends on the type:
- **Integration tests** (test executables): `cmd/tests/`
- **Unit tests** (Go test files): Same directory as code (`*_test.go`)
- **E2E tests**: `test/` or `cmd/tests/e2e/`

### Q: What's the difference between `internal/` and `pkg/`?

**A**:
- `internal/`: Cannot be imported by external projects (Go enforces this)
- `pkg/`: Can be imported by anyone
- Rule: Start with `internal/`, promote to `pkg/` only when needed

### Q: Can `cmd/` import from `internal/`?

**A**: Yes! That's the pattern:
- `cmd/` contains minimal startup code
- `cmd/` imports and orchestrates `internal/` packages
- `internal/` contains the real logic

### Q: Where do vendor dependencies go?

**A**:
- Use Go modules (`go.mod`, `go.sum`)
- Optional: `vendor/` for vendoring (add to `.gitignore`)
- Never store dependencies in `pkg/` or `internal/`

---

## Migration Checklist

Moving to this structure from flat directory:

- [x] Create directory structure
- [x] Move test files to `cmd/tests/`
- [x] Move configs to `config/`
- [x] Move docs to `docs/`
- [x] Move scripts to `scripts/`
- [ ] Extract `gateway.go` to `internal/gateway/`
- [ ] Create `pkg/` packages for reusable code
- [ ] Add examples in `examples/`
- [ ] Update import paths
- [ ] Update documentation
- [ ] Test everything still works

---

## References

### Go Project Structure
- [Organize Like a Pro: Go Project Folder Structures](https://medium.com/@smart_byte_labs/organize-like-a-pro-a-simple-guide-to-go-project-folder-structures-e85e9c1769c2)
- [11 Tips for Structuring Go Projects - Alex Edwards](https://www.alexedwards.net/blog/11-tips-for-structuring-your-go-projects)
- [Optimal Project Layout for Large-Scale Go Applications](https://leapcell.io/blog/optimal-project-layout-for-large-scale-go-applications)
- [Best Practices for Go Project Structure](https://medium.com/@nandoseptian/best-practices-for-go-project-structure-and-code-organization-486898990d0a)
- [Go Project Structure Best Practices - TutorialEdge](https://tutorialedge.net/golang/go-project-structure-best-practices/)

### General Best Practices
- [Folder Structure Best Practices: Complete Guide](https://compresto.app/blog/folder-structure-best-practices)
- [Guide to Folder Structure & Best Practices](https://www.suitefiles.com/guide/the-guide-to-folder-structures-best-practices-for-professional-service-firms-and-more/)

---

**Remember**: This structure is a foundation. Adapt it to your specific needs, but always prioritize:

1. **Clarity** - Anyone should understand the organization
2. **Scalability** - Easy to add new features/components
3. **Maintainability** - Easy to find and modify code
4. **Separation** - Clear boundaries between concerns
