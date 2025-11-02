# Database Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/nabbar/golib/database)

**Comprehensive database management with GORM ORM integration and generic Key-Value store abstraction.**

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for testing, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [GORM Package](#gorm-package)
- [Key-Value Packages](#key-value-packages)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [Related Documentation](#related-documentation)
- [License](#license)

---

## Overview

The **database** package collection provides two complementary database paradigms:

1. **GORM Package** - Relational database support with ORM
2. **KV Packages** - Generic Key-Value store abstraction

### Design Philosophy

- **Abstraction**: Clean interfaces hiding implementation complexity
- **Type Safety**: Leverage Go generics for compile-time safety
- **Performance**: Efficient connection pooling and caching
- **Observability**: Built-in monitoring and health checks
- **Flexibility**: Support multiple databases and storage backends

---

## Key Features

### GORM Package

| Feature | Description |
|---------|-------------|
| **Multi-DB Support** | MySQL, PostgreSQL, SQLite, SQL Server |
| **Connection Pooling** | Configurable pool settings (idle, max, lifetime) |
| **Logger Integration** | golib/logger integration with log levels |
| **Context Support** | Full context.Context for cancellation/deadlines |
| **Monitoring** | Health checks and connection status |
| **Config Validation** | Struct validation with go-playground/validator |
| **Thread-Safe** | Safe for concurrent use |

### Key-Value Packages

| Package | Purpose |
|---------|---------|
| **kvtypes** | Type definitions and interfaces |
| **kvdriver** | Database driver abstraction |
| **kvitem** | Individual key-value item operations |
| **kvmap** | Map-based operations |
| **kvtable** | Table-level operations |

---

## Architecture

### Package Structure

```
database/
├── gorm/                    # GORM ORM integration
│   ├── config.go           # Configuration structures
│   ├── driver.go           # Database driver detection
│   ├── interface.go        # Public interfaces
│   ├── model.go            # Implementation
│   ├── monitor.go          # Health monitoring
│   └── errors.go           # Error definitions
│
└── kv*/                     # Key-Value packages
    ├── kvtypes/            # Type definitions
    ├── kvdriver/           # Driver abstraction
    ├── kvitem/             # Item operations
    ├── kvmap/              # Map operations
    └── kvtable/            # Table operations
```

### GORM Architecture

```
┌─────────────────────────────────────────────────────┐
│              GORM Package                            │
│                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────┐ │
│  │   Config     │  │   Monitor    │  │  Logger  │ │
│  │  Validation  │  │  Health Check│  │ Integr.  │ │
│  └──────────────┘  └──────────────┘  └──────────┘ │
│         │                  │                │       │
│         ▼                  ▼                ▼       │
│  ┌──────────────────────────────────────────────┐  │
│  │         GORM Interface                       │  │
│  │  (Connection Pool, Transaction, Migration)   │  │
│  └──────────────────────────────────────────────┘  │
│                        │                             │
│                        ▼                             │
│  ┌──────────────────────────────────────────────┐  │
│  │      Database Drivers                        │  │
│  │  MySQL │ PostgreSQL │ SQLite │ SQL Server   │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Key-Value Packages Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                    KV Packages Ecosystem                          │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                     kvtypes                              │    │
│  │  (Generic Types [K comparable, V any])                  │    │
│  │  - Interfaces                                            │    │
│  │  - Type Definitions                                      │    │
│  │  - Common Structures                                     │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                    │
│                              ▼                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    kvdriver                              │    │
│  │  (Database Driver Abstraction)                          │    │
│  │  - Connection Management                                │    │
│  │  - Generic Operations [K, V]                            │    │
│  │  - Driver Interface                                     │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                    │
│          ┌───────────────────┼───────────────────┐              │
│          ▼                   ▼                   ▼              │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐       │
│  │   kvitem     │   │    kvmap     │   │   kvtable    │       │
│  │              │   │              │   │              │       │
│  │ Single Item  │   │  Map-based   │   │   Table      │       │
│  │ Operations:  │   │ Operations:  │   │ Operations:  │       │
│  │              │   │              │   │              │       │
│  │ - Get        │   │ - GetAll     │   │ - Schema     │       │
│  │ - Set        │   │ - SetAll     │   │ - Namespace  │       │
│  │ - Delete     │   │ - Delete     │   │ - Create     │       │
│  │ - Exists     │   │ - Keys       │   │ - Drop       │       │
│  │ - Expire     │   │ - Values     │   │ - List       │       │
│  │              │   │ - Range      │   │ - Migrate    │       │
│  └──────────────┘   └──────────────┘   └──────────────┘       │
│          │                   │                   │              │
│          └───────────────────┼───────────────────┘              │
│                              ▼                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              Storage Backend                             │    │
│  │  (Redis, BadgerDB, BoltDB, etc.)                        │    │
│  └─────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────┘

Data Flow:
  Application → kvitem/kvmap/kvtable → kvdriver → Storage Backend
  
Type Safety:
  Compile-time type checking through Go generics [K comparable, V any]
  
Abstraction Layers:
  1. kvtypes:  Generic type system
  2. kvdriver: Storage abstraction
  3. kvitem/kvmap/kvtable: Operation-specific interfaces
```

---

## GORM Package

### Supported Databases

| Database | Driver | CGO Required | Status |
|----------|--------|--------------|--------|
| MySQL | gorm.io/driver/mysql | No | ✅ Supported |
| PostgreSQL | gorm.io/driver/postgres | No | ✅ Supported |
| SQLite | gorm.io/driver/sqlite | **Yes** | ✅ Supported |
| SQL Server | gorm.io/driver/sqlserver | No | ✅ Supported |

### Installation

```bash
go get github.com/nabbar/golib/database/gorm

# Install database driver
go get -u gorm.io/driver/mysql        # MySQL
go get -u gorm.io/driver/postgres     # PostgreSQL
go get -u gorm.io/driver/sqlite       # SQLite (requires CGO)
go get -u gorm.io/driver/sqlserver    # SQL Server
```

### Configuration

```go
type Config struct {
    Driver               Driver          // Database driver (MySQL, PostgreSQL, etc.)
    Name                 string          // Connection name/identifier
    DSN                  string          // Data Source Name
    
    // Connection Pool
    EnableConnectionPool bool            // Enable connection pooling
    PoolMaxIdleConns    int              // Max idle connections
    PoolMaxOpenConns    int              // Max open connections  
    PoolMaxConnLifetime time.Duration    // Max connection lifetime
    PoolMaxConnIdleTime time.Duration    // Max connection idle time
    
    // GORM Options
    SkipDefaultTransaction bool          // Skip default transaction
    PrepareStmt           bool           // Prepare statements
    DisableAutomaticPing  bool           // Disable automatic ping
    
    // Logger
    LogLevel              logger.LogLevel // GORM log level
    LogThreshold          time.Duration   // Slow query threshold
}
```

### Quick Start

```go
import (
    "log"
    libgorm "github.com/nabbar/golib/database/gorm"
)

func main() {
    // Configure database
    cfg := &libgorm.Config{
        Driver: libgorm.DriverMysql,
        Name:   "mydb",
        DSN:    "user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4",
        EnableConnectionPool: true,
        PoolMaxIdleConns:     10,
        PoolMaxOpenConns:     100,
    }
    
    // Create database instance
    db, err := libgorm.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Get GORM DB
    gormDB := db.GetDB()
    
    // Use GORM as usual
    type User struct {
        ID   uint
        Name string
    }
    gormDB.AutoMigrate(&User{})
    gormDB.Create(&User{Name: "Alice"})
}
```

### Monitoring

```go
import (
    libgorm "github.com/nabbar/golib/database/gorm"
    montps "github.com/nabbar/golib/monitor/types"
)

// Register health check
func setupMonitoring(db libgorm.Database, pool montps.Pool) {
    monitor := db.Monitor(context.Background())
    
    pool.RegisterPool("database", func(ctx context.Context) (bool, error) {
        return monitor.IsRunning(), nil
    })
}
```

### Context Management

```go
// With context cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Query with context
var users []User
result := db.GetDB().WithContext(ctx).Find(&users)
if result.Error != nil {
    log.Printf("Query error: %v", result.Error)
}
```

---

## Key-Value Packages

The KV packages provide a generic, type-safe abstraction for key-value storage.

### Package Overview

**kvtypes** - Core type definitions:
- Generic types with `[K comparable, V any]`
- Interface definitions
- Common structures

**kvdriver** - Driver abstraction:
- Database driver interface
- Connection management
- Generic operations

**kvitem** - Item operations:
- Single key-value operations
- Get, Set, Delete
- Expiration support

**kvmap** - Map operations:
- Bulk operations
- Map-based interface
- Batch processing

**kvtable** - Table operations:
- Table-level abstractions
- Namespace support
- Schema management

### Example

```go
import (
    "github.com/nabbar/golib/database/kvdriver"
    "github.com/nabbar/golib/database/kvitem"
)

// Create driver (implementation-specific)
driver := kvdriver.New(config)

// Use items
item := kvitem.New[string, User](driver, "users")
item.Set(ctx, "user:123", User{ID: 123, Name: "Alice"})

user, err := item.Get(ctx, "user:123")
if err != nil {
    log.Fatal(err)
}
```

---

## Performance

### GORM Package

**Connection Pooling:**
- **Idle Connections**: Reused for new queries (minimal overhead)
- **Max Open Connections**: Prevents database overload
- **Connection Lifetime**: Automatic recycling of stale connections

**Benchmarks (MySQL):**

| Operation | Time | Notes |
|-----------|------|-------|
| Simple Query | ~500µs | With connection pool |
| Insert | ~1ms | Single record |
| Batch Insert (100) | ~50ms | Using CreateInBatches |
| Transaction | ~2ms | Begin/Commit overhead |
| Connection Open | ~5ms | First connection (one-time) |

*Benchmarks on localhost MySQL 8.0, AMD64*

### Key-Value Packages

**Generic Overhead:**
- Zero runtime overhead (generics compiled away)
- Type safety at compile time
- Inlining of generic functions

---

## Use Cases

### GORM Package

**Web Applications**
```go
// User management API
func GetUser(c *gin.Context) {
    db := c.MustGet("db").(libgorm.Database)
    
    var user User
    result := db.GetDB().First(&user, c.Param("id"))
    if result.Error != nil {
        c.JSON(404, gin.H{"error": "user not found"})
        return
    }
    
    c.JSON(200, user)
}
```

**Multi-Tenant Applications**
```go
// Separate databases per tenant
func getTenantDB(tenantID string) (libgorm.Database, error) {
    cfg := &libgorm.Config{
        Driver: libgorm.DriverPostgres,
        DSN:    fmt.Sprintf("host=localhost user=%s dbname=tenant_%s", user, tenantID),
    }
    return libgorm.New(cfg)
}
```

**Migrations**
```go
// Database migrations
db.GetDB().AutoMigrate(&User{}, &Product{}, &Order{})
```

### KV Packages

**Session Storage**
```go
sessions := kvitem.New[string, Session](driver, "sessions")
sessions.Set(ctx, sessionID, sessionData)
```

**Caching Layer**
```go
cache := kvitem.New[string, CachedData](driver, "cache")
cache.SetWithExpiration(ctx, key, data, 5*time.Minute)
```

---

## Best Practices

### GORM

1. **Use Connection Pooling**
```go
cfg.EnableConnectionPool = true
cfg.PoolMaxIdleConns = 10
cfg.PoolMaxOpenConns = 100
```

2. **Set Connection Lifetimes**
```go
cfg.PoolMaxConnLifetime = time.Hour
cfg.PoolMaxConnIdleTime = 10 * time.Minute
```

3. **Use Contexts**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
db.GetDB().WithContext(ctx).Find(&users)
```

4. **Handle Errors**
```go
result := db.GetDB().Create(&user)
if result.Error != nil {
    log.Printf("Error: %v", result.Error)
}
```

5. **Close Connections**
```go
defer db.Close()
```

### KV Packages

1. **Use Generics for Type Safety**
```go
// Type-safe
users := kvitem.New[string, User](driver, "users")

// Avoid any
items := kvitem.New[string, any](driver, "data") // Less safe
```

2. **Handle Context Cancellation**
```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
item.Set(ctx, key, value)
```

---

## Testing

Comprehensive testing documentation is available in [TESTING.md](TESTING.md).

**Quick Test:**
```bash
cd database/gorm
go test -v -cover
```

**With SQLite (requires CGO):**
```bash
CGO_ENABLED=1 go test -v -cover
```

**Test Metrics:**
- GORM: 41 test specifications, 61.6% coverage (with CGO)
- Ginkgo v2 + Gomega framework
- Multi-database testing

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass existing tests
- Maintain or improve test coverage
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized
- Document all public APIs with GoDoc

**Testing**
- Write tests for all new features
- Test with multiple databases (MySQL, PostgreSQL, SQLite)
- Verify connection pool behavior
- Include error cases

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**GORM Package**
- Connection pool metrics and monitoring dashboard
- Automatic failover and replica support
- Query performance profiling
- Schema versioning and migration tools
- Multi-database transaction support

**KV Packages**
- Redis driver implementation
- Bolt/BadgerDB driver implementation
- TTL and expiration policies
- Distributed locking
- Pub/Sub support

**General**
- Performance benchmarking suite
- Visual query analyzer
- Automatic index suggestions
- Database comparison tools

Suggestions and contributions are welcome via GitHub issues.

---

## Related Documentation

### External Libraries
- **[GORM](https://gorm.io/)** - The fantastic ORM library for Go
- **[GORM Drivers](https://gorm.io/docs/connecting_to_the_database.html)** - Database driver documentation
- **[go-playground/validator](https://github.com/go-playground/validator)** - Struct validation
- **[Ginkgo Testing](https://github.com/onsi/ginkgo)** - BDD testing framework

### Related Golib Packages
- **[logger](../logger/README.md)** - Logging integration
- **[monitor](../monitor/README.md)** - Health monitoring
- **[config](../config/README.md)** - Configuration management
- **[context](../context/README.md)** - Context management

### Database Documentation
- **[MySQL](https://dev.mysql.com/doc/)** - MySQL official documentation
- **[PostgreSQL](https://www.postgresql.org/docs/)** - PostgreSQL documentation
- **[SQLite](https://www.sqlite.org/docs.html)** - SQLite documentation
- **[SQL Server](https://docs.microsoft.com/en-us/sql/)** - SQL Server documentation

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2019 Nicolas JUHEL

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/database)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **GORM Documentation**: [gorm.io](https://gorm.io/)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Source Code**: [GitHub Repository](https://github.com/nabbar/golib)

---

*This package is part of the [golib](https://github.com/nabbar/golib) project.*
