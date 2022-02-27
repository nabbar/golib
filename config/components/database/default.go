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
	"bytes"
	"encoding/json"
	"time"

	libcfg "github.com/nabbar/golib/config"
	libdbs "github.com/nabbar/golib/database"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	libsts "github.com/nabbar/golib/status"
	spfcbr "github.com/spf13/cobra"
	spfvbr "github.com/spf13/viper"
)

var _defaultConfig = []byte(`{
  "driver": "",
  "name": "",
  "dsn": "",
  "skip-default-transaction": false,
  "full-save-associations": false,
  "dry-run": false,
  "prepare-stmt": false,
  "disable-automatic-ping": false,
  "disable-foreign-key-constraint-when-migrating": false,
  "disable-nested-transaction": false,
  "allow-global-update": false,
  "query-fields": false,
  "create-batch-size": 0,
  "enable-connection-pool": false,
  "pool-max-idle-conns": 0,
  "pool-max-open-conns": 0,
  "pool-conn-max-lifetime": "0s",
  "disabled": false,
  "status": ` + string(libsts.DefaultConfig(libcfg.JSONIndent)) + `
}`)

func SetDefaultConfig(cfg []byte) {
	_defaultConfig = cfg
}

func DefaultConfig(indent string) []byte {
	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, _defaultConfig, indent, libcfg.JSONIndent); err != nil {
		return _defaultConfig
	} else {
		return res.Bytes()
	}
}

func (c *componentDatabase) DefaultConfig(indent string) []byte {
	return DefaultConfig(indent)
}

func (c *componentDatabase) RegisterFlag(Command *spfcbr.Command, Viper *spfvbr.Viper) error {
	_ = Command.PersistentFlags().String(c.key+".driver", "", "driver to use for the DSN. It must be one of mysql, psql, sqlite, sqlserver, clickhouse.")
	_ = Command.PersistentFlags().String(c.key+".name", "", "string to identify the instance into status.")
	_ = Command.PersistentFlags().String(c.key+".dsn", "", "string options to connect to database following database engine. See https://gorm.io/docs/connecting_to_the_database.html for more information.")
	_ = Command.PersistentFlags().Bool(c.key+".skip-default-transaction", false, "disable the default transaction for single create, update, delete operations.")
	_ = Command.PersistentFlags().Bool(c.key+".full-save-associations", false, "full save associations")
	_ = Command.PersistentFlags().Bool(c.key+".dry-run", false, "generate sql without execute")
	_ = Command.PersistentFlags().Bool(c.key+".prepare-stmt", false, "executes the given query in cached statement")
	_ = Command.PersistentFlags().Bool(c.key+".disable-automatic-ping", false, "is used to disable the automatic ping to the database server")
	_ = Command.PersistentFlags().Bool(c.key+".disable-foreign-key-constraint-when-migrating", false, "is used to disable the foreign key constraint when migrating the database")
	_ = Command.PersistentFlags().Bool(c.key+".disable-nested-transaction", false, "disable nested transaction")
	_ = Command.PersistentFlags().Bool(c.key+".allow-global-update", false, "allow global update")
	_ = Command.PersistentFlags().Bool(c.key+".query-fields", false, "executes the SQL query with all fields of the table")
	_ = Command.PersistentFlags().Int(c.key+".create-batch-size", 0, "default create batch size")
	_ = Command.PersistentFlags().Bool(c.key+".enable-connection-pool", false, "is used to create a connection pool")
	_ = Command.PersistentFlags().Int(c.key+".pool-max-idle-conns", 0, "sets the maximum number of connections idle in the connection pool if enable")
	_ = Command.PersistentFlags().Int(c.key+".pool-max-open-conns", 0, "sets the maximum number of connections open in the connection pool if enable")
	_ = Command.PersistentFlags().Duration(c.key+".pool-conn-max-lifetime", 5*time.Minute, "sets the maximum lifetime of connections in the connection pool if enable")
	_ = Command.PersistentFlags().Bool(c.key+".disabled", false, "allow to disable a database connection without clean his configuration")

	if err := Viper.BindPFlag(c.key+".driver", Command.PersistentFlags().Lookup(c.key+".driver")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".name", Command.PersistentFlags().Lookup(c.key+".name")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".dsn", Command.PersistentFlags().Lookup(c.key+".dsn")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".skip-default-transaction", Command.PersistentFlags().Lookup(c.key+".skip-default-transaction")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".full-save-associations", Command.PersistentFlags().Lookup(c.key+".full-save-associations")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".dry-run", Command.PersistentFlags().Lookup(c.key+".dry-run")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".prepare-stmt", Command.PersistentFlags().Lookup(c.key+".prepare-stmt")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".disable-automatic-ping", Command.PersistentFlags().Lookup(c.key+".disable-automatic-ping")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".disable-foreign-key-constraint-when-migrating", Command.PersistentFlags().Lookup(c.key+".disable-foreign-key-constraint-when-migrating")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".disable-nested-transaction", Command.PersistentFlags().Lookup(c.key+".disable-nested-transaction")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".allow-global-update", Command.PersistentFlags().Lookup(c.key+".allow-global-update")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".query-fields", Command.PersistentFlags().Lookup(c.key+".query-fields")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".create-batch-size", Command.PersistentFlags().Lookup(c.key+".create-batch-size")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".enable-connection-pool", Command.PersistentFlags().Lookup(c.key+".enable-connection-pool")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".pool-max-idle-conns", Command.PersistentFlags().Lookup(c.key+".pool-max-idle-conns")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".pool-max-open-conns", Command.PersistentFlags().Lookup(c.key+".pool-max-open-conns")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".pool-conn-max-lifetime", Command.PersistentFlags().Lookup(c.key+".pool-conn-max-lifetime")); err != nil {
		return err
	} else if err = Viper.BindPFlag(c.key+".disabled", Command.PersistentFlags().Lookup(c.key+".disabled")); err != nil {
		return err
	} else if err = libsts.RegisterFlag(c.key+".status", Command, Viper); err != nil {
		return err
	}

	return nil
}

func (c *componentDatabase) _getConfig(getCfg libcfg.FuncComponentConfigGet) (libdbs.Config, liberr.Error) {
	var (
		cnf = libdbs.Config{}
		vpr = c.vpr()
	)

	if !c._CheckDep() {
		return cnf, ErrorComponentNotInitialized.Error(nil)
	}

	if err := getCfg(c.key, &cnf); err != nil {
		return cnf, ErrorParamsInvalid.Error(err)
	}

	fct := func() liblog.Logger {
		if l, e := c._GetLogger(); e != nil {
			return liblog.GetDefault()
		} else {
			return l
		}
	}

	cnf.RegisterLogger(fct, c.li, c.ls)
	cnf.RegisterContext(c.ctx)

	if val := vpr.GetString(c.key + ".driver"); val != "" {
		cnf.Driver = libdbs.DriverFromString(val)
	}
	if val := vpr.GetString(c.key + ".name"); val != "" {
		cnf.Name = val
	}
	if val := vpr.GetString(c.key + ".dsn"); val != "" {
		cnf.DSN = val
	}
	if val := vpr.GetBool(c.key + ".skip-default-transaction"); val {
		cnf.SkipDefaultTransaction = true
	}
	if val := vpr.GetBool(c.key + ".full-save-associations"); val {
		cnf.FullSaveAssociations = true
	}
	if val := vpr.GetBool(c.key + ".dry-run"); val {
		cnf.DryRun = true
	}
	if val := vpr.GetBool(c.key + ".prepare-stmt"); val {
		cnf.PrepareStmt = true
	}
	if val := vpr.GetBool(c.key + ".disable-automatic-ping"); val {
		cnf.DisableAutomaticPing = true
	}
	if val := vpr.GetBool(c.key + ".disable-foreign-key-constraint-when-migrating"); val {
		cnf.DisableForeignKeyConstraintWhenMigrating = true
	}
	if val := vpr.GetBool(c.key + ".disable-nested-transaction"); val {
		cnf.DisableNestedTransaction = true
	}
	if val := vpr.GetBool(c.key + ".allow-global-update"); val {
		cnf.AllowGlobalUpdate = true
	}
	if val := vpr.GetBool(c.key + ".query-fields"); val {
		cnf.QueryFields = true
	}
	if val := vpr.GetInt(c.key + ".create-batch-size"); val != 0 {
		cnf.CreateBatchSize = val
	}
	if val := vpr.GetBool(c.key + ".enable-connection-pool"); val {
		cnf.EnableConnectionPool = true
	}
	if val := vpr.GetInt(c.key + ".pool-max-idle-conns"); val != 0 {
		cnf.PoolMaxIdleConns = val
	}
	if val := vpr.GetInt(c.key + ".pool-max-open-conns"); val != 0 {
		cnf.PoolMaxOpenConns = val
	}
	if val := vpr.GetDuration(c.key + ".pool-conn-max-lifetime"); val != 0 {
		cnf.PoolConnMaxLifetime = val
	}
	if val := vpr.GetBool(c.key + ".disabled"); val {
		cnf.Disabled = true
	}

	if err := cnf.Validate(); err != nil {
		return cnf, ErrorConfigInvalid.Error(err)
	}

	return cnf, nil
}
