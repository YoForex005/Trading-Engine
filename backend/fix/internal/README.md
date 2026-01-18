# Internal Packages

This directory contains private application code that cannot be imported by external projects.

## Structure

```
internal/
├── gateway/        # FIX gateway core implementation
├── session/        # Session lifecycle management
└── message/        # FIX message handlers
```

## Purpose

The `internal/` directory is a special Go convention that prevents external packages from importing this code. This allows us to:
- Refactor freely without breaking external dependencies
- Enforce encapsulation at the language level
- Keep implementation details private

## Planned Modules

### `gateway/`
Core FIX gateway functionality:
- Connection management
- Message routing
- Session orchestration
- LP integration

### `session/`
Session lifecycle management:
- Logon/logout handling
- Heartbeat management
- Sequence number tracking
- Session state persistence

### `message/`
FIX message processing:
- Message parsing and validation
- Message construction
- Protocol-specific handlers
- Field extraction utilities

## Migration Status

**Status**: Directory structure created, awaiting refactoring

**Next Steps**:
1. Extract gateway logic from root `gateway.go`
2. Create session management module
3. Implement message handlers
4. Add unit tests for each module
