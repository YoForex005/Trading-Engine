# Public Packages

This directory contains reusable packages that can be imported by external projects.

## Structure

```
pkg/
├── fixclient/      # Reusable FIX client library
└── types/          # Shared types and interfaces
```

## Purpose

Code in `pkg/` is designed to be:
- **Reusable** - Can be imported by other projects
- **Stable** - Has well-defined public APIs
- **Well-documented** - Includes comprehensive documentation
- **Tested** - Has thorough test coverage

## Planned Modules

### `fixclient/`
Reusable FIX client library:
- High-level FIX client interface
- Connection management
- Message sending/receiving
- Common operations (subscribe, place order, etc.)

**Example Usage**:
```go
import "trading-engine/backend/fix/pkg/fixclient"

client := fixclient.New(config)
client.Connect()
client.SubscribeMarketData("EURUSD")
```

### `types/`
Shared types and constants:
- Order types and enums
- Quote structures
- Position types
- FIX field constants
- Common interfaces

**Example Usage**:
```go
import "trading-engine/backend/fix/pkg/types"

order := types.Order{
    Symbol: "EURUSD",
    Side: types.Buy,
    Quantity: 10000,
}
```

## API Stability

Since `pkg/` code is public and importable:
- Breaking changes require major version bumps
- Deprecated features should be marked clearly
- Maintain backward compatibility when possible
- Document all exported functions and types

## Migration Status

**Status**: Directory structure created, awaiting implementation

**Next Steps**:
1. Define core types in `types/`
2. Extract reusable logic into `fixclient/`
3. Add comprehensive documentation
4. Create examples showing usage
5. Add integration tests
