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

package database

import (
	"time"

	libvpr "github.com/nabbar/golib/viper"

	"github.com/nabbar/golib/database/gorm"

	spfcbr "github.com/spf13/cobra"
	spfvpr "github.com/spf13/viper"
)

func (o *componentDatabase) RegisterFlag(Command *spfcbr.Command) error {
	var (
		key string
		vpr *spfvpr.Viper
	)

	if vpr = o._getSPFViper(); vpr == nil {
		return ErrorComponentNotInitialized.Error(nil)
	} else if key = o._getKey(); len(key) < 1 {
		return ErrorComponentNotInitialized.Error(nil)
	}

	_ = Command.PersistentFlags().String(key+".driver", "", "driver to use for the DSN. It must be one of mysql, psql, sqlite, sqlserver, clickhouse.")
	_ = Command.PersistentFlags().String(key+".name", "", "string to identify the instance into status.")
	_ = Command.PersistentFlags().String(key+".dsn", "", "string options to connect to database following database engine. See https://gorm.io/docs/connecting_to_the_database.html for more information.")
	_ = Command.PersistentFlags().Bool(key+".skip-default-transaction", false, "disable the default transaction for single create, update, delete operations.")
	_ = Command.PersistentFlags().Bool(key+".full-save-associations", false, "full save associations")
	_ = Command.PersistentFlags().Bool(key+".dry-run", false, "generate sql without execute")
	_ = Command.PersistentFlags().Bool(key+".prepare-stmt", false, "executes the given query in cached statement")
	_ = Command.PersistentFlags().Bool(key+".disable-automatic-ping", false, "is used to disable the automatic ping to the database server")
	_ = Command.PersistentFlags().Bool(key+".disable-foreign-key-constraint-when-migrating", false, "is used to disable the foreign key constraint when migrating the database")
	_ = Command.PersistentFlags().Bool(key+".disable-nested-transaction", false, "disable nested transaction")
	_ = Command.PersistentFlags().Bool(key+".allow-global-update", false, "allow global update")
	_ = Command.PersistentFlags().Bool(key+".query-fields", false, "executes the SQL query with all fields of the table")
	_ = Command.PersistentFlags().Int(key+".create-batch-size", 0, "default create batch size")
	_ = Command.PersistentFlags().Bool(key+".enable-connection-pool", false, "is used to create a connection pool")
	_ = Command.PersistentFlags().Int(key+".pool-max-idle-conns", 0, "sets the maximum number of connections idle in the connection pool if enable")
	_ = Command.PersistentFlags().Int(key+".pool-max-open-conns", 0, "sets the maximum number of connections open in the connection pool if enable")
	_ = Command.PersistentFlags().Duration(key+".pool-conn-max-lifetime", 5*time.Minute, "sets the maximum lifetime of connections in the connection pool if enable")
	_ = Command.PersistentFlags().Bool(key+".disabled", false, "allow to disable a database connection without clean his configuration")

	if err := vpr.BindPFlag(key+".driver", Command.PersistentFlags().Lookup(key+".driver")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".name", Command.PersistentFlags().Lookup(key+".name")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".dsn", Command.PersistentFlags().Lookup(key+".dsn")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".skip-default-transaction", Command.PersistentFlags().Lookup(key+".skip-default-transaction")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".full-save-associations", Command.PersistentFlags().Lookup(key+".full-save-associations")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".dry-run", Command.PersistentFlags().Lookup(key+".dry-run")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".prepare-stmt", Command.PersistentFlags().Lookup(key+".prepare-stmt")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".disable-automatic-ping", Command.PersistentFlags().Lookup(key+".disable-automatic-ping")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".disable-foreign-key-constraint-when-migrating", Command.PersistentFlags().Lookup(key+".disable-foreign-key-constraint-when-migrating")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".disable-nested-transaction", Command.PersistentFlags().Lookup(key+".disable-nested-transaction")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".allow-global-update", Command.PersistentFlags().Lookup(key+".allow-global-update")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".query-fields", Command.PersistentFlags().Lookup(key+".query-fields")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".create-batch-size", Command.PersistentFlags().Lookup(key+".create-batch-size")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".enable-connection-pool", Command.PersistentFlags().Lookup(key+".enable-connection-pool")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".pool-max-idle-conns", Command.PersistentFlags().Lookup(key+".pool-max-idle-conns")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".pool-max-open-conns", Command.PersistentFlags().Lookup(key+".pool-max-open-conns")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".pool-conn-max-lifetime", Command.PersistentFlags().Lookup(key+".pool-conn-max-lifetime")); err != nil {
		return err
	} else if err = vpr.BindPFlag(key+".disabled", Command.PersistentFlags().Lookup(key+".disabled")); err != nil {
		return err
	}

	return nil
}

func (o *componentDatabase) _getConfig() (*gorm.Config, error) {
	var (
		key string
		cfg gorm.Config
		vpr libvpr.Viper
		err error
	)

	if vpr = o._getViper(); vpr == nil {
		return nil, ErrorComponentNotInitialized.Error(nil)
	} else if key = o._getKey(); len(key) < 1 {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	if e := vpr.UnmarshalKey(key, &cfg); e != nil {
		return nil, ErrorParamInvalid.Error(e)
	}

	cfg.RegisterLogger(o.getLogger, o.li, o.ls)
	cfg.RegisterContext(o.x.GetContext)

	if val := vpr.GetString(key + ".driver"); val != "" {
		cfg.Driver = gorm.DriverFromString(val)
	}
	if val := vpr.GetString(key + ".name"); val != "" {
		cfg.Name = val
	}
	if val := vpr.GetString(key + ".dsn"); val != "" {
		cfg.DSN = val
	}
	if val := vpr.GetBool(key + ".skip-default-transaction"); val {
		cfg.SkipDefaultTransaction = true
	}
	if val := vpr.GetBool(key + ".full-save-associations"); val {
		cfg.FullSaveAssociations = true
	}
	if val := vpr.GetBool(key + ".dry-run"); val {
		cfg.DryRun = true
	}
	if val := vpr.GetBool(key + ".prepare-stmt"); val {
		cfg.PrepareStmt = true
	}
	if val := vpr.GetBool(key + ".disable-automatic-ping"); val {
		cfg.DisableAutomaticPing = true
	}
	if val := vpr.GetBool(key + ".disable-foreign-key-constraint-when-migrating"); val {
		cfg.DisableForeignKeyConstraintWhenMigrating = true
	}
	if val := vpr.GetBool(key + ".disable-nested-transaction"); val {
		cfg.DisableNestedTransaction = true
	}
	if val := vpr.GetBool(key + ".allow-global-update"); val {
		cfg.AllowGlobalUpdate = true
	}
	if val := vpr.GetBool(key + ".query-fields"); val {
		cfg.QueryFields = true
	}
	if val := vpr.GetInt(key + ".create-batch-size"); val != 0 {
		cfg.CreateBatchSize = val
	}
	if val := vpr.GetBool(key + ".enable-connection-pool"); val {
		cfg.EnableConnectionPool = true
	}
	if val := vpr.GetInt(key + ".pool-max-idle-conns"); val != 0 {
		cfg.PoolMaxIdleConns = val
	}
	if val := vpr.GetInt(key + ".pool-max-open-conns"); val != 0 {
		cfg.PoolMaxOpenConns = val
	}
	if val := vpr.GetDuration(key + ".pool-conn-max-lifetime"); val != 0 {
		cfg.PoolConnMaxLifetime = val
	}
	if val := vpr.GetBool(key + ".disabled"); val {
		cfg.Disabled = true
	}

	if err = cfg.Validate(); err != nil {
		return nil, ErrorConfigInvalid.Error(err)
	}

	return &cfg, nil
}
