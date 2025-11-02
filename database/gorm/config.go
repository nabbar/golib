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

package gorm

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	libval "github.com/go-playground/validator/v10"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	loggrm "github.com/nabbar/golib/logger/gorm"
	moncfg "github.com/nabbar/golib/monitor/types"
	gormdb "gorm.io/gorm"
	gorcls "gorm.io/gorm/clause"
	gorlog "gorm.io/gorm/logger"
)

type Config struct {
	// Driver is the driver to use for the DSN. It must be one of mysql, psql, sqlite, sqlserver, clickhouse.
	Driver Driver `json:"driver" yaml:"driver" toml:"driver" mapstructure:"driver"`

	// Name is a string to identify the instance into status.
	Name string `json:"name" yaml:"name" toml:"name" mapstructure:"name"`

	// DSN is the string options to connect to database following database engine. See https://gorm.io/docs/connecting_to_the_database.html for more information.
	DSN string `json:"dsn" yaml:"dsn" toml:"dsn" mapstructure:"dsn"`

	// SkipDefaultTransaction disable the default transaction for single create, update, delete operations.
	// This single transactions by default is to ensure database data integrity.
	SkipDefaultTransaction bool `json:"skip-default-transaction" yaml:"skip-default-transaction" toml:"skip-default-transaction" mapstructure:"skip-default-transaction"`

	// FullSaveAssociations full save associations.
	FullSaveAssociations bool `json:"full-save-associations" yaml:"full-save-associations" toml:"full-save-associations" mapstructure:"full-save-associations"`

	// DryRun generate sql without execute.
	DryRun bool `json:"dry-run" yaml:"dry-run" toml:"dry-run" mapstructure:"dry-run"`

	// PrepareStmt executes the given query in cached statement.
	PrepareStmt bool `json:"prepare-stmt" yaml:"prepare-stmt" toml:"prepare-stmt" mapstructure:"prepare-stmt"`

	// DisableAutomaticPing is used to disable the automatic ping to the database server.
	DisableAutomaticPing bool `json:"disable-automatic-ping" yaml:"disable-automatic-ping" toml:"disable-automatic-ping" mapstructure:"disable-automatic-ping"`

	// DisableForeignKeyConstraintWhenMigrating is used to disable the foreign key constraint when migrating the database.
	DisableForeignKeyConstraintWhenMigrating bool `json:"disable-foreign-key-constraint-when-migrating" yaml:"disable-foreign-key-constraint-when-migrating" toml:"disable-foreign-key-constraint-when-migrating" mapstructure:"disable-foreign-key-constraint-when-migrating"`

	// DisableNestedTransaction disable nested transaction.
	DisableNestedTransaction bool `json:"disable-nested-transaction" yaml:"disable-nested-transaction" toml:"disable-nested-transaction" mapstructure:"disable-nested-transaction"`

	// AllowGlobalUpdate allow global update.
	AllowGlobalUpdate bool `json:"allow-global-update" yaml:"allow-global-update" toml:"allow-global-update" mapstructure:"allow-global-update"`

	// QueryFields executes the SQL query with all fields of the table.
	QueryFields bool `json:"query-fields" yaml:"query-fields" toml:"query-fields" mapstructure:"query-fields"`

	// CreateBatchSize default create batch size.
	CreateBatchSize int `json:"create-batch-size" yaml:"create-batch-size" toml:"create-batch-size" mapstructure:"create-batch-size"`

	// EnableConnectionPool is used to create a connection pool.
	EnableConnectionPool bool `json:"enable-connection-pool" yaml:"enable-connection-pool" toml:"enable-connection-pool" mapstructure:"enable-connection-pool"`

	// PoolMaxIdleConns sets the maximum number of connections idle in the connection pool.
	PoolMaxIdleConns int `json:"pool-max-idle-conns" yaml:"pool-max-idle-conns" toml:"pool-max-idle-conns" mapstructure:"pool-max-idle-conns"`

	// PoolMaxOpenConns sets the maximum number of connections open in the connection pool.
	PoolMaxOpenConns int `json:"pool-max-open-conns" yaml:"pool-max-open-conns" toml:"pool-max-open-conns" mapstructure:"pool-max-open-conns"`

	// PoolConnMaxLifetime sets the maximum lifetime of connections in the connection pool.
	PoolConnMaxLifetime time.Duration `json:"pool-conn-max-lifetime" yaml:"pool-conn-max-lifetime" toml:"pool-conn-max-lifetime" mapstructure:"pool-conn-max-lifetime"`

	// Disabled allow to disable a database connection without clean his configuration.
	Disabled bool `mapstructure:"disabled" json:"disabled" yaml:"disabled" toml:"disabled"`

	// Monitor defined the monitoring configuration
	Monitor moncfg.Config `mapstructure:"monitor" json:"monitor" yaml:"monitor" toml:"monitor"`

	//@TODO : implement logger options with new logger

	ctx  context.Context
	flog func() gorlog.Interface
}

// Validate allow checking if the config' struct is valid with the awaiting model
func (c *Config) Validate() liberr.Error {
	var e = ErrorValidatorError.Error(nil)

	if err := libval.New().Struct(c); err != nil {
		if er, ok := err.(*libval.InvalidValidationError); ok {
			e.Add(er)
		}

		for _, er := range err.(libval.ValidationErrors) {
			//nolint #goerr113
			e.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", er.Namespace(), er.ActualTag()))
		}
	}

	if !e.HasParent() {
		e = nil
	}

	return e
}

func (c *Config) RegisterLogger(fct func() liblog.Logger, ignoreRecordNotFoundError bool, slowThreshold time.Duration) {
	c.flog = func() gorlog.Interface {
		return loggrm.New(fct, ignoreRecordNotFoundError, slowThreshold)
	}
}

func (c *Config) RegisterGORMLogger(fct FuncGormLog) {
	c.flog = fct
}

func (c *Config) RegisterContext(fct context.Context) {
	c.ctx = fct
}

func (c *Config) Config() *gormdb.Config {
	cfg := &gormdb.Config{
		SkipDefaultTransaction:                   c.SkipDefaultTransaction,
		NamingStrategy:                           nil,
		FullSaveAssociations:                     c.FullSaveAssociations,
		Logger:                                   nil,
		NowFunc:                                  nil,
		DryRun:                                   c.DryRun,
		PrepareStmt:                              c.PrepareStmt,
		DisableAutomaticPing:                     c.DisableAutomaticPing,
		DisableForeignKeyConstraintWhenMigrating: c.DisableForeignKeyConstraintWhenMigrating,
		DisableNestedTransaction:                 c.DisableNestedTransaction,
		AllowGlobalUpdate:                        c.AllowGlobalUpdate,
		QueryFields:                              c.QueryFields,
		CreateBatchSize:                          c.CreateBatchSize,
		ClauseBuilders:                           make(map[string]gorcls.ClauseBuilder),
		ConnPool:                                 nil,
		Dialector:                                c.Driver.Dialector(c.DSN),
		Plugins:                                  make(map[string]gormdb.Plugin),
	}

	if c.flog != nil {
		cfg.Logger = c.flog()
	}

	return cfg
}

func (c *Config) New(cfg *gormdb.Config) (*gormdb.DB, liberr.Error) {
	if cfg == nil {
		cfg = c.Config()
	}

	o, e := gormdb.Open(c.Driver.Dialector(c.DSN), cfg)

	if e != nil {
		return nil, ErrorDatabaseOpen.Error(e)
	}

	if c.ctx != nil {
		o.WithContext(c.ctx)
	}

	if c.EnableConnectionPool {
		var db *sql.DB

		if db, e = o.DB(); e != nil {
			return nil, ErrorDatabaseOpenPool.Error(e)
		}

		if c.PoolMaxIdleConns > 0 {
			db.SetMaxIdleConns(c.PoolMaxIdleConns)
		}

		if c.PoolMaxOpenConns > 0 {
			db.SetMaxOpenConns(c.PoolMaxOpenConns)
		}

		if c.PoolConnMaxLifetime > 0 {
			db.SetConnMaxLifetime(c.PoolConnMaxLifetime)
		}
	}

	return o, nil
}
