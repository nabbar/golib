# Database Package - Testing Documentation

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![Ginkgo](https://img.shields.io/badge/Ginkgo-v2-green)](https://github.com/onsi/ginkgo)

Comprehensive testing guide for the database package using Ginkgo v2/Gomega BDD framework.

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for test generation, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [GORM Package Testing](#gorm-package-testing)
- [KV Packages Testing](#kv-packages-testing)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Database Setup](#database-setup)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)
- [Contributing](#contributing)

---

## Overview

The database package features comprehensive testing across GORM integration and Key-Value packages.

### Test Metrics

| Package | Specs | Coverage | Notes |
|---------|-------|----------|-------|
| **gorm** | 41 | 61.6% | Requires CGO for SQLite |
| **kvdriver** | 25+ | >85% | Generic driver tests |
| **Total** | 65+ | >70% | Combined |

---

## Test Framework

### Ginkgo v2 + Gomega

**Features:**
- BDD-style test organization
- Parallel execution support
- Rich assertion library
- Database lifecycle management

**Installation:**
```bash
go get github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega
```

---

## GORM Package Testing

### Test Organization

```
gorm/
├── gorm_suite_test.go       # Suite setup
├── config_test.go           # Configuration tests (15 specs)
├── database_test.go         # Database operations (26 specs)
└── drivers/                 # Driver-specific tests
```

### Running GORM Tests

**Quick Test (in-memory SQLite):**
```bash
cd database/gorm
CGO_ENABLED=1 go test -v
```

**With Coverage:**
```bash
CGO_ENABLED=1 go test -v -cover
CGO_ENABLED=1 go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Using Ginkgo:**
```bash
CGO_ENABLED=1 ginkgo -v
CGO_ENABLED=1 ginkgo -v -cover
```

### Test Scenarios

#### 1. **Configuration Tests** (config_test.go)

**Scenarios:**
- Configuration validation
- DSN parsing
- Driver selection
- Pool settings
- GORM options

**Example:**
```go
Describe("Configuration", func() {
    It("should validate required fields", func() {
        cfg := &Config{
            Driver: DriverMysql,
            Name:   "test",
            DSN:    "",
        }
        
        err := cfg.Validate()
        Expect(err).To(HaveOccurred())
    })
    
    It("should accept valid configuration", func() {
        cfg := &Config{
            Driver: DriverSqlite,
            Name:   "test",
            DSN:    "file::memory:?cache=shared",
        }
        
        err := cfg.Validate()
        Expect(err).NotTo(HaveOccurred())
    })
})
```

#### 2. **Database Operations** (database_test.go)

**Scenarios:**
- Connection establishment
- Connection pooling
- Query execution
- Transaction management
- Connection lifecycle
- Error handling

**Example:**
```go
Describe("Database Operations", func() {
    var db Database
    
    BeforeEach(func() {
        cfg := &Config{
            Driver: DriverSqlite,
            DSN:    "file::memory:?cache=shared",
        }
        var err error
        db, err = New(cfg)
        Expect(err).NotTo(HaveOccurred())
    })
    
    AfterEach(func() {
        db.Close()
    })
    
    It("should execute queries", func() {
        type User struct {
            ID   uint
            Name string
        }
        
        gormDB := db.GetDB()
        gormDB.AutoMigrate(&User{})
        
        result := gormDB.Create(&User{Name: "Test"})
        Expect(result.Error).NotTo(HaveOccurred())
        Expect(result.RowsAffected).To(Equal(int64(1)))
    })
})
```

### Multi-Database Testing

Test against multiple databases:

```go
var _ = Describe("Multi-Database Tests", func() {
    databases := []struct {
        name   string
        driver Driver
        dsn    string
        skip   bool
    }{
        {
            name:   "SQLite",
            driver: DriverSqlite,
            dsn:    "file::memory:?cache=shared",
            skip:   false,
        },
        {
            name:   "MySQL",
            driver: DriverMysql,
            dsn:    os.Getenv("MYSQL_DSN"),
            skip:   os.Getenv("MYSQL_DSN") == "",
        },
        {
            name:   "PostgreSQL",
            driver: DriverPostgres,
            dsn:    os.Getenv("POSTGRES_DSN"),
            skip:   os.Getenv("POSTGRES_DSN") == "",
        },
    }
    
    for _, db := range databases {
        Context(db.name, func() {
            if db.skip {
                Skip("Database not configured")
            }
            
            It("should connect", func() {
                cfg := &Config{
                    Driver: db.driver,
                    DSN:    db.dsn,
                }
                database, err := New(cfg)
                Expect(err).NotTo(HaveOccurred())
                defer database.Close()
            })
        })
    }
})
```

---

## KV Packages Testing

### Test Organization

```
kvdriver/
├── driver_suite_test.go     # Suite setup
└── driver_test.go           # Driver tests (25+ specs)
```

### Running KV Tests

```bash
cd database/kvdriver
go test -v -cover
```

### Test Scenarios

**Generic Type Tests:**
```go
Describe("Generic Types", func() {
    It("should work with string keys", func() {
        driver := NewDriver[string, int]()
        // Test operations
    })
    
    It("should work with custom types", func() {
        type UserID int
        driver := NewDriver[UserID, User]()
        // Test operations
    })
})
```

---

## Test Coverage

### Coverage by Component

| Component | File | Coverage | Notes |
|-----------|------|----------|-------|
| Configuration | config.go | ~90% | Validation logic |
| Database | model.go | ~70% | Core operations |
| Monitor | monitor.go | ~60% | Health checks |
| Driver | driver.go | ~50% | OS-specific |
| Errors | errors.go | ~80% | Error definitions |

**Overall Coverage**: 61.6% (GORM with CGO)

### Coverage Gaps

- **Driver detection**: OS-specific code paths
- **Error scenarios**: Some edge cases
- **Connection failures**: Network-dependent tests

---

## Database Setup

### SQLite (In-Memory)

**No setup required:**
```go
cfg := &Config{
    Driver: DriverSqlite,
    DSN:    "file::memory:?cache=shared",
}
```

**Requires:** `CGO_ENABLED=1`

### MySQL (Docker)

```bash
docker run --name mysql-test \
    -e MYSQL_ROOT_PASSWORD=root \
    -e MYSQL_DATABASE=testdb \
    -p 3306:3306 \
    -d mysql:8.0

export MYSQL_DSN="root:root@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True"
```

### PostgreSQL (Docker)

```bash
docker run --name postgres-test \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=testdb \
    -p 5432:5432 \
    -d postgres:15

export POSTGRES_DSN="host=localhost user=postgres password=postgres dbname=testdb port=5432 sslmode=disable"
```

### SQL Server (Docker)

```bash
docker run --name sqlserver-test \
    -e 'ACCEPT_EULA=Y' \
    -e 'SA_PASSWORD=YourStrong!Passw0rd' \
    -p 1433:1433 \
    -d mcr.microsoft.com/mssql/server:2019-latest

export SQLSERVER_DSN="sqlserver://sa:YourStrong!Passw0rd@localhost:1433?database=testdb"
```

---

## Best Practices

### 1. Use BeforeEach/AfterEach

```go
var db Database

BeforeEach(func() {
    cfg := &Config{
        Driver: DriverSqlite,
        DSN:    "file::memory:?cache=shared",
    }
    var err error
    db, err = New(cfg)
    Expect(err).NotTo(HaveOccurred())
})

AfterEach(func() {
    db.Close()
})
```

### 2. Clean Up Resources

```go
It("should clean up connections", func() {
    db, _ := New(cfg)
    defer db.Close()
    
    // Test operations
    
    // Verify cleanup
    Expect(db.IsRunning()).To(BeFalse())
})
```

### 3. Test Error Cases

```go
It("should handle invalid DSN", func() {
    cfg := &Config{
        Driver: DriverMysql,
        DSN:    "invalid-dsn",
    }
    
    _, err := New(cfg)
    Expect(err).To(HaveOccurred())
})
```

### 4. Use Context Timeouts

```go
It("should respect context timeout", func() {
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    // Long-running query
    result := db.GetDB().WithContext(ctx).Raw("SELECT SLEEP(1)").Scan(&res)
    Expect(result.Error).To(HaveOccurred())
})
```

### 5. Test Connection Pool

```go
It("should manage connection pool", func() {
    cfg := &Config{
        Driver:               DriverSqlite,
        DSN:                  "file::memory:?cache=shared",
        EnableConnectionPool: true,
        PoolMaxOpenConns:     10,
        PoolMaxIdleConns:     5,
    }
    
    db, err := New(cfg)
    Expect(err).NotTo(HaveOccurred())
    defer db.Close()
    
    // Test concurrent connections
})
```

---

## CI/CD Integration

### GitHub Actions

```yaml
test-database:
  runs-on: ubuntu-latest
  
  services:
    mysql:
      image: mysql:8.0
      env:
        MYSQL_ROOT_PASSWORD: root
        MYSQL_DATABASE: testdb
      ports:
        - 3306:3306
      options: --health-cmd="mysqladmin ping" --health-interval=10s
    
    postgres:
      image: postgres:15
      env:
        POSTGRES_PASSWORD: postgres
        POSTGRES_DB: testdb
      ports:
        - 5432:5432
      options: --health-cmd pg_isready --health-interval 10s
  
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Test GORM
      env:
        CGO_ENABLED: 1
        MYSQL_DSN: "root:root@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True"
        POSTGRES_DSN: "host=localhost user=postgres password=postgres dbname=testdb sslmode=disable"
      run: |
        cd database/gorm
        go test -v -race -cover
```

### GitLab CI

```yaml
test-database:
  services:
    - mysql:8.0
    - postgres:15
  
  variables:
    MYSQL_ROOT_PASSWORD: root
    MYSQL_DATABASE: testdb
    POSTGRES_PASSWORD: postgres
    POSTGRES_DB: testdb
    CGO_ENABLED: 1
  
  script:
    - cd database/gorm
    - go test -v -race -cover
  
  coverage: '/coverage: \d+\.\d+% of statements/'
```

---

## Contributing

When adding new features:

1. **Write tests first** (TDD approach)
2. **Cover all databases** (MySQL, PostgreSQL, SQLite, SQL Server)
3. **Test error cases** (invalid config, connection failures)
4. **Test connection pool** behavior
5. **Update coverage** metrics
6. **Document test scenarios**

### Test Template

```go
var _ = Describe("New Feature", func() {
    var db Database
    
    BeforeEach(func() {
        cfg := &Config{
            Driver: DriverSqlite,
            DSN:    "file::memory:?cache=shared",
        }
        var err error
        db, err = New(cfg)
        Expect(err).NotTo(HaveOccurred())
    })
    
    AfterEach(func() {
        db.Close()
    })
    
    Describe("Feature behavior", func() {
        It("should handle basic case", func() {
            // Test implementation
            Expect(result).To(Equal(expected))
        })
        
        Context("when error occurs", func() {
            It("should return error", func() {
                // Test error handling
                Expect(err).To(HaveOccurred())
            })
        })
    })
})
```

---

## Support

For issues or questions:

- **Test Failures**: Check output with `-v` flag
- **CGO Issues**: Ensure `CGO_ENABLED=1` for SQLite
- **Database Setup**: Verify connection strings
- **Feature Questions**: See [README.md](README.md)
- **Bug Reports**: [GitHub Issues](https://github.com/nabbar/golib/issues)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

*Part of the [golib](https://github.com/nabbar/golib) testing suite.*
