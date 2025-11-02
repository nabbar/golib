/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

// Package gorm provides a comprehensive wrapper around GORM ORM with connection management,
// monitoring, logging integration, and configuration validation.
//
// Features:
//   - Multiple database support (MySQL, PostgreSQL, SQLite, SQL Server)
//   - Connection pooling with configurable settings
//   - Logger integration with golib/logger
//   - Context management for cancellation and deadlines
//   - Health monitoring and connection status
//   - Configuration validation with go-playground/validator
//   - Thread-safe concurrent access
//
// Example usage:
//
//	import libgorm "github.com/nabbar/golib/database/gorm"
//
//	cfg := &libgorm.Config{
//	    Driver: libgorm.DriverMysql,
//	    Name:   "mydb",
//	    DSN:    "user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4",
//	    EnableConnectionPool: true,
//	    PoolMaxIdleConns:     10,
//	    PoolMaxOpenConns:     100,
//	}
//
//	db, err := libgorm.New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	gormDB := db.GetDB()
//	gormDB.Create(&User{Name: "Alice"})
package gorm

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	gormdb "gorm.io/gorm"
	gorlog "gorm.io/gorm/logger"
)

// FuncGormLog is a function type that returns a GORM logger interface.
// Used for providing custom GORM loggers to the database instance.
type FuncGormLog func() gorlog.Interface

// Database is the main interface for database operations.
// It wraps a GORM DB instance with additional features like monitoring,
// logging, and connection management.
type Database interface {
	// GetDB returns the underlying GORM DB instance.
	// Use this to access all GORM functionality directly.
	GetDB() *gormdb.DB

	// SetDb replaces the underlying GORM DB instance.
	// Use with caution as this affects all concurrent operations.
	SetDb(db *gormdb.DB)

	// Close closes the database connection and releases resources.
	// Safe to call multiple times.
	Close()

	// WaitNotify blocks until the context is cancelled or database closes.
	// Used for graceful shutdown and connection lifecycle management.
	WaitNotify(ctx context.Context, cancel context.CancelFunc)

	// CheckConn verifies the database connection is alive.
	// Returns an error if the connection is not functional.
	CheckConn() liberr.Error

	// Config returns the GORM configuration used by this database.
	Config() *gormdb.Config

	// RegisterContext registers a context function for the database.
	// The context is used for cancellation and deadlines.
	RegisterContext(fct context.Context)

	// RegisterLogger registers a golib logger for database operations.
	// Parameters:
	//   - fct: Function returning a logger instance
	//   - ignoreRecordNotFoundError: If true, record not found errors are not logged
	//   - slowThreshold: Threshold for slow query logging
	RegisterLogger(fct func() liblog.Logger, ignoreRecordNotFoundError bool, slowThreshold time.Duration)

	// RegisterGORMLogger registers a GORM-specific logger.
	// Use this for custom GORM logging behavior.
	RegisterGORMLogger(fct func() gorlog.Interface)

	// Monitor creates a monitor for health checking and status reporting.
	// Returns a Monitor instance that can be registered with a monitor pool.
	Monitor(vrs libver.Version) (montps.Monitor, error)
}

// New creates a new Database instance with the given configuration.
// The configuration is validated before creating the database.
//
// Parameters:
//   - cfg: Database configuration including driver, DSN, and pool settings
//
// Returns:
//   - Database: Configured database instance
//   - liberr.Error: Error if configuration is invalid or connection fails
//
// Example:
//
//	cfg := &Config{
//	    Driver: DriverPostgres,
//	    DSN:    "host=localhost user=postgres dbname=mydb sslmode=disable",
//	}
//	db, err := New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
func New(cfg *Config) (Database, liberr.Error) {
	if d, e := cfg.New(nil); e != nil {
		return nil, e
	} else {
		v := new(atomic.Value)
		v.Store(d)

		c := new(atomic.Value)
		c.Store(cfg)

		return &database{
			m: sync.Mutex{},
			v: v,
			c: c,
		}, nil
	}
}
