# golib database

This package provides helpers for database management, including a GORM helper and a generic Key-Value (KV) database abstraction. It is designed to simplify database operations and provide a unified interface for both relational and KV stores.

## Sub-packages Overview

### 1. `gorm`

- **Purpose:** Helper for [GORM](https://gorm.io/) ORM.
- **Features:**
    - Configuration struct for GORM options (transactions, connection pool, logger, etc.).
    - Database initialization and health check.
    - Context and logger registration.
    - Monitoring integration.
- **Usage:**  
  Use the `Config` struct to configure your GORM connection, then use the `New` function to instantiate a database connection.  
  Example:
  ```go
  import "github.com/nabbar/golib/database/gorm"

  cfg := &gorm.Config{ /* ... */ }
  db, err := gorm.New(cfg)
  ```
- See [gorm Subpackage](#gorm-subpackage) for more details.

### 2. `kvdriver`

- **Purpose:** Generic driver interface for Key-Value databases.
- **Features:**
    - Define comparison functions for keys (equality, contains, empty).
    - Abstracts the basic operations: Get, Set, Delete, List, Search, Walk.
    - Allows custom driver implementations by providing function pointers.
- **Usage:**  
  Implement the required functions and use `kvdriver.New` to create a driver instance.
- See [kvdriver Subpackage](#kvdriver-subpackage) for more details.

### 3. `kvmap`

- **Purpose:** Implements a KV driver using a `map[comparable]any` transformation.
- **Features:**
    - Serializes/deserializes models to/from `map[comparable]any` using JSON.
    - Useful for stores where data is naturally represented as maps.
- **Usage:**  
  Use `kvmap.New` to create a driver, providing serialization logic and storage functions.
- See [kvmap Subpackage](#kvmap-subpackage) for more details.

### 4. `kvtable`

- **Purpose:** Table abstraction for KV stores, providing higher-level operations.
- **Features:**
    - Operations: Get, Delete, List, Search, Walk.
    - Each table is backed by a KV driver.
- **Usage:**  
  Create a table with `kvtable.New(driver)` and use the table interface for record management.
- See [kvtable Subpackage](#kvtable-subpackage) for more details.

### 5. `kvitem`

- **Purpose:** Represents a single record/item in a KV table.
- **Features:**
    - Load, store, remove, and clean a record.
    - Change detection between loaded and stored state.
- **Usage:**  
  Use `kvitem.New(driver, key)` to create an item, then use its methods to manipulate the record.
- See [kvitem Subpackage](#kvitem-subpackage) for more details.

### 6. `kvtypes`

- **Purpose:** Defines generic types and interfaces for KV drivers, items, and tables.
- **Features:**
    - `KVDriver`: Interface for drivers.
    - `KVItem`: Interface for single records.
    - `KVTable`: Interface for table operations.
    - Type aliases for walk functions.

---

## Example Usage

```go
import (
    "github.com/nabbar/golib/database/kvdriver"
    "github.com/nabbar/golib/database/kvtable"
    "github.com/nabbar/golib/database/kvitem"
)

// Define your key and model types
type Key string
type Model struct { /* ... */ }

// Implement required functions for your storage backend
// ...

// Create a driver
driver := kvdriver.New(compare, newFunc, getFunc, setFunc, delFunc, listFunc, searchFunc, walkFunc)

// Create a table
table := kvtable.New(driver)

// Get an item
item, err := table.Get("myKey")
if err == nil {
    model := item.Get()
    // ...
}
```

---

## `kvdriver` Subpackage

The `kvdriver` package provides a generic, extensible driver interface for Key-Value (KV) databases. It allows you to define custom drivers by supplying function pointers for all core operations, and supports advanced key comparison logic.

### Features

- Generic driver interface for any key (`comparable`) and model type
- Customizable comparison functions for keys (equality, contains, empty)
- Abstracts basic KV operations: Get, Set, Delete, List, Search, Walk
- Error codes and messages integrated with the `liberr` package
- Optional support for advanced search and walk operations

---

### Main Types & Functions

#### Comparison Functions

Define how keys are compared and matched:

- `CompareEqual[K]`: func(ref, part K) bool
- `CompareContains[K]`: func(ref, part K) bool
- `CompareEmpty[K]`: func(part K) bool

Create a comparison instance:

```go
cmp := kvdriver.NewCompare(eqFunc, containsFunc, emptyFunc)
```

#### Driver Functions

Implement the following function types for your backend:

- `FuncNew[K, M]`: Create a new driver instance
- `FuncGet[K, M]`: Get a model by key
- `FuncSet[K, M]`: Set a model by key
- `FuncDel[K]`: Delete a model by key
- `FuncList[K]`: List all keys
- `FuncSearch[K]`: (Optional) Search keys by pattern
- `FuncWalk[K, M]`: (Optional) Walk through all records

#### Creating a Driver

Instantiate a driver by providing all required functions:

```go
driver := kvdriver.New(
    cmp,      // Compare[K]
    newFunc,  // FuncNew[K, M]
    getFunc,  // FuncGet[K, M]
    setFunc,  // FuncSet[K, M]
    delFunc,  // FuncDel[K]
    listFunc, // FuncList[K]
    searchFunc, // FuncSearch[K] (optional)
    walkFunc,   // FuncWalk[K, M] (optional)
)
```

The returned driver implements the `KVDriver[K, M]` interface from `kvtypes`.

---

### Example

```go
import (
    "github.com/nabbar/golib/database/kvdriver"
    "github.com/nabbar/golib/database/kvtypes"
)

type Key string
type Model struct { /* ... */ }

// Define comparison functions
eq := func(a, b Key) bool { return a == b }
contains := func(a, b Key) bool { return strings.Contains(string(a), string(b)) }
empty := func(a Key) bool { return a == "" }
cmp := kvdriver.NewCompare(eq, contains, empty)

// Implement storage functions (get, set, etc.)
// ...

driver := kvdriver.New(cmp, newFunc, getFunc, setFunc, delFunc, listFunc, nil, nil)
```

---

### Error Handling

All errors are returned as `liberr.Error` with specific codes (e.g., `ErrorBadInstance`, `ErrorGetFunction`). Always check errors after each operation.

---

### Notes

- The package is fully generic and requires Go 1.18+.
- If `FuncSearch` or `FuncWalk` are not provided, default implementations are used (based on `List` and `Get`).
- Integrates with the `kvtypes` package for type definitions.

---

Voici une documentation en anglais pour le package `github.com/nabbar/golib/database/kvmap`, à inclure dans votre `README.md` ou documentation technique.

---

## `kvmap` Subpackage

The `kvmap` package provides a generic Key-Value (KV) driver implementation using a `map[comparable]any` as the underlying storage. It serializes and deserializes models to and from maps using JSON, making it suitable for stores where data is naturally represented as maps.

### Features

- Generic driver for any key (`comparable`) and model type
- Serializes/deserializes models to `map[comparable]any` via JSON
- Customizable storage, retrieval, and management functions
- Supports advanced operations: Get, Set, Delete, List, Search, Walk
- Integrates with the `kvtypes` and `kvdriver` packages

---

### Main Types & Functions

#### Function Types

- `FuncNew[K, M]`: Creates a new driver instance
- `FuncGet[K, MK]`: Retrieves a map for a given key
- `FuncSet[K, MK]`: Stores a map for a given key
- `FuncDel[K]`: Deletes a key
- `FuncList[K]`: Lists all keys
- `FuncSearch[K]`: (Optional) Searches keys by prefix/pattern
- `FuncWalk[K, M]`: (Optional) Walks through all records

#### Creating a Driver

Instantiate a driver by providing all required functions:

```go
import "github.com/nabbar/golib/database/kvmap"

driver := kvmap.New(
    cmp,      // Compare[K]
    newFunc,  // FuncNew[K, M]
    getFunc,  // FuncGet[K, MK]
    setFunc,  // FuncSet[K, MK]
    delFunc,  // FuncDel[K]
    listFunc, // FuncList[K]
    searchFunc, // FuncSearch[K] (optional)
    walkFunc,   // FuncWalk[K, M] (optional)
)
```

The returned driver implements the `KVDriver[K, M]` interface.

---

### Example

```go
import (
    "github.com/nabbar/golib/database/kvmap"
    "github.com/nabbar/golib/database/kvdriver"
)

type Key string
type Model struct { /* ... */ }

// Define comparison functions
cmp := kvdriver.NewCompare(eqFunc, containsFunc, emptyFunc)

// Implement storage functions (get, set, etc.)
// ...

driver := kvmap.New(cmp, newFunc, getFunc, setFunc, delFunc, listFunc, nil, nil)
```

---

### Error Handling

All errors are returned as `liberr.Error` with specific codes (e.g., `ErrorBadInstance`, `ErrorGetFunction`). Always check errors after each operation.

---

### Notes

- The package is fully generic and requires Go 1.18+.
- If `FuncSearch` or `FuncWalk` are not provided, default implementations are used (based on `List` and `Get`).
- Serialization and deserialization use JSON under the hood.
- Integrates with the `kvtypes` and `kvdriver` packages for type definitions and comparison logic.

---

Voici une documentation en anglais pour le package `github.com/nabbar/golib/database/kvtable`, à inclure dans votre `README.md` ou documentation technique.

---

## `kvtable` Subpackage

The `kvtable` package provides a high-level table abstraction for Key-Value (KV) stores, built on top of a generic KV driver. It simplifies record management by exposing table-like operations and returning item wrappers for each record.

### Features

- Table abstraction for any key (`comparable`) and model type
- Operations: Get, Delete, List, Search, Walk
- Each table is backed by a pluggable KV driver
- Returns `KVItem` wrappers for each record, enabling further manipulation
- Integrates with the `kvtypes` and `kvitem` packages

---

### Main Types & Functions

#### Creating a Table

Instantiate a table by providing a KV driver:

```go
import (
    "github.com/nabbar/golib/database/kvtable"
    "github.com/nabbar/golib/database/kvtypes"
)

driver := /* your KVDriver[K, M] implementation */
table := kvtable.New(driver)
```

#### Table Operations

- `Get(key K) (KVItem[K, M], error)`: Retrieve an item by key.
- `Del(key K) error`: Delete an item by key.
- `List() ([]KVItem[K, M], error)`: List all items in the table.
- `Search(pattern K) ([]KVItem[K, M], error)`: Search items by pattern.
- `Walk(fct FuncWalk[K, M]) error`: Walk through all items, applying a function.

Example usage:

```go
item, err := table.Get("myKey")
if err == nil {
    model := item.Get()
    // ...
}
```

---

### Error Handling

All operations return errors as `liberr.Error` with specific codes (e.g., `ErrorBadDriver`). Always check errors after each operation.

---

### Notes

- The package is fully generic and requires Go 1.18+.
- Each table instance is backed by a driver implementing the `KVDriver` interface.
- Returned `KVItem` wrappers allow further operations like load, store, and change detection.

---

Voici une documentation en anglais pour le package `github.com/nabbar/golib/database/kvitem`, à inclure dans votre `README.md` ou documentation technique.

---

## `kvitem` Subpackage

The `kvitem` package provides a generic wrapper for managing a single record (item) in a Key-Value (KV) store. It is designed to work with any key and model type, and relies on a pluggable KV driver for storage operations. The package offers change detection, atomic state management, and error handling.

### Features

- Generic item wrapper for any key (`comparable`) and model type
- Load, store, remove, and clean a record
- Change detection between loaded and stored state
- Atomic operations for thread safety
- Integrates with the `kvtypes` package for driver abstraction
- Custom error codes and messages

---

### Main Types & Functions

#### Creating an Item

Instantiate a new item by providing a KV driver and a key:

```go
import (
    "github.com/nabbar/golib/database/kvitem"
    "github.com/nabbar/golib/database/kvtypes"
)

driver := /* your KVDriver[K, M] implementation */
item := kvitem.New(driver, key)
```

#### Item Operations

- `Set(model M)`: Set the model value to be stored.
- `Get() M`: Get the current model (from store or last loaded).
- `Load() error`: Load the model from the backend.
- `Store(force bool) error`: Store the model to the backend. If `force` is true, always writes even if unchanged.
- `Remove() error`: Remove the item from the backend.
- `Clean()`: Reset the loaded and stored model to zero value.
- `HasChange() bool`: Returns true if the stored model differs from the loaded model.
- `Key() K`: Returns the item's key.

#### Example Usage

```go
item := kvitem.New(driver, "user:123")
err := item.Load()
if err == nil {
    model := item.Get()
    // modify model...
    item.Set(model)
    if item.HasChange() {
        _ = item.Store(false)
    }
}
```

---

### Error Handling

All operations return errors as `liberr.Error` with specific codes (e.g., `ErrorLoadFunction`, `ErrorStoreFunction`). Always check errors after each operation.

---

### Notes

- The package is fully generic and requires Go 1.18+.
- Uses atomic values for thread-safe state management.
- Integrates with the `kvtypes` package for driver and item interfaces.
- Change detection is based on deep equality of loaded and stored models.

---

Voici une documentation en anglais pour le package `github.com/nabbar/golib/database/gorm`, à inclure dans votre `README.md` ou documentation technique.

---

## `gorm` Subpackage

The `gorm` package provides helpers for configuring, initializing, and managing [GORM](https://gorm.io/) database connections in Go applications. It offers advanced configuration options, context and logger integration, connection pooling, and monitoring support.

### Features

- Configuration struct for all GORM options (driver, DSN, transactions, pooling, etc.)
- Database initialization and health checking
- Context and logger registration
- Monitoring integration for health and metrics
- Error codes and messages for robust error handling
- Support for multiple database drivers (MySQL, PostgreSQL, SQLite, SQL Server, ClickHouse)

---

### Main Types & Functions

#### `Config` Struct

Defines all configuration options for a GORM database connection:

- `Driver`: Database driver (`mysql`, `psql`, `sqlite`, `sqlserver`, `clickhouse`)
- `Name`: Instance name for status/monitoring
- `DSN`: Data Source Name (connection string)
- Transaction, pooling, and migration options
- Monitoring configuration

#### Example Usage

```go
import (
    "github.com/nabbar/golib/database/gorm"
)

cfg := &gorm.Config{
    Driver: gorm.DriverMysql,
    Name:   "main-db",
    DSN:    "user:pass@tcp(localhost:3306)/dbname",
    EnableConnectionPool: true,
    PoolMaxIdleConns: 5,
    PoolMaxOpenConns: 10,
    // ... other options
}

db, err := gorm.New(cfg)
if err != nil {
    // handle error
}
defer db.Close()
```

#### Registering Logger and Context

```go
import (
    "github.com/nabbar/golib/logger"
    "github.com/nabbar/golib/context"
)

cfg.RegisterLogger(func() logger.Logger { /* ... */ }, false, 200*time.Millisecond)
cfg.RegisterContext(func() context.Context { /* ... */ })
```

#### Monitoring

You can enable monitoring and health checks for your database instance:

```go
mon, err := db.Monitor(version)
if err != nil {
    // handle error
}
```

#### Error Handling

All errors are returned as `liberr.Error` with specific codes (e.g., `ErrorDatabaseOpen`, `ErrorValidatorError`). Always check errors after each operation.

---

### Supported Drivers

- MySQL
- PostgreSQL
- SQLite
- SQL Server
- ClickHouse

Select the driver using the `Driver` field in the config.

---

### Notes

- The package is thread-safe and uses atomic values for state management.
- Integrates with the `liberr`, `liblog`, and `libctx` packages for error, logging, and context management.
- Monitoring is based on the `monitor` subpackage and can be customized via the config.

---
