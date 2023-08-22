//go:build !arm && !arm64
// +build !arm,!arm64

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
	"strings"

	drvclk "gorm.io/driver/clickhouse"
	drvmys "gorm.io/driver/mysql"
	drvpsq "gorm.io/driver/postgres"
	drvsql "gorm.io/driver/sqlite"
	drvsrv "gorm.io/driver/sqlserver"
	gormdb "gorm.io/gorm"
)

const (
	DriverNone       = ""
	DriverMysql      = "mysql"
	DriverPostgreSQL = "psql"
	DriverSQLite     = "sqlite"
	DriverSQLServer  = "sqlserver"
	DriverClikHouse  = "clickhouse"
)

type Driver string

func DriverFromString(drv string) Driver {
	switch strings.ToLower(drv) {

	case strings.ToLower(DriverMysql):
		return DriverMysql

	case strings.ToLower(DriverPostgreSQL):
		return DriverPostgreSQL

	case strings.ToLower(DriverSQLite):
		return DriverSQLite

	case strings.ToLower(DriverSQLServer):
		return DriverSQLServer

	case strings.ToLower(DriverClikHouse):
		return DriverClikHouse

	default:
		return DriverNone
	}
}

func (d Driver) String() string {
	return string(d)
}

func (d Driver) Dialector(dsn string) gormdb.Dialector {
	switch d {

	case DriverMysql:
		return drvmys.Open(dsn)

	case DriverPostgreSQL:
		return drvpsq.Open(dsn)

	case DriverSQLite:
		return drvsql.Open(dsn)

	case DriverSQLServer:
		return drvsrv.Open(dsn)

	case DriverClikHouse:
		return drvclk.Open(dsn)

	default:
		return nil
	}
}
