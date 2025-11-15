/*
MIT License

Copyright (c) 2022 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package gorm_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libgorm "github.com/nabbar/golib/database/gorm"
)

var _ = Describe("GORM Config", func() {
	Describe("Driver", func() {
		It("should convert string to Driver type", func() {
			Expect(libgorm.DriverFromString("mysql")).To(Equal(libgorm.Driver("mysql")))
			Expect(libgorm.DriverFromString("psql")).To(Equal(libgorm.Driver("psql")))
			Expect(libgorm.DriverFromString("sqlite")).To(Equal(libgorm.Driver("sqlite")))
			Expect(libgorm.DriverFromString("sqlserver")).To(Equal(libgorm.Driver("sqlserver")))
		})

		It("should be case insensitive", func() {
			Expect(libgorm.DriverFromString("MYSQL")).To(Equal(libgorm.Driver("mysql")))
			Expect(libgorm.DriverFromString("MySQL")).To(Equal(libgorm.Driver("mysql")))
			Expect(libgorm.DriverFromString("SQLite")).To(Equal(libgorm.Driver("sqlite")))
		})

		It("should return DriverNone for unknown driver", func() {
			Expect(libgorm.DriverFromString("unknown")).To(Equal(libgorm.Driver("")))
			Expect(libgorm.DriverFromString("")).To(Equal(libgorm.Driver("")))
		})

		It("should convert Driver to string", func() {
			Expect(libgorm.Driver("mysql").String()).To(Equal("mysql"))
			Expect(libgorm.Driver("psql").String()).To(Equal("psql"))
			Expect(libgorm.Driver("sqlite").String()).To(Equal("sqlite"))
			Expect(libgorm.Driver("sqlserver").String()).To(Equal("sqlserver"))
		})

		It("should create valid dialectors", func() {
			mysqlDriver := libgorm.DriverFromString("mysql")
			Expect(mysqlDriver.Dialector("test-dsn")).ToNot(BeNil())

			sqliteDriver := libgorm.DriverFromString("sqlite")
			Expect(sqliteDriver.Dialector(":memory:")).ToNot(BeNil())
		})

		It("should return nil dialector for invalid driver", func() {
			invalidDriver := libgorm.Driver("")
			Expect(invalidDriver.Dialector("test")).To(BeNil())
		})
	})

	Describe("Config Creation", func() {
		It("should create a valid config for SQLite", func() {
			cfg := &libgorm.Config{
				Driver: libgorm.DriverSQLite,
				Name:   "test-db",
				DSN:    ":memory:",
			}
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Driver).To(Equal(libgorm.Driver(libgorm.DriverSQLite)))
			Expect(cfg.Name).To(Equal("test-db"))
			Expect(cfg.DSN).To(Equal(":memory:"))
		})

		It("should handle config with connection pool settings", func() {
			cfg := &libgorm.Config{
				Driver:               libgorm.DriverSQLite,
				DSN:                  ":memory:",
				EnableConnectionPool: true,
				PoolMaxIdleConns:     10,
				PoolMaxOpenConns:     100,
			}
			Expect(cfg.EnableConnectionPool).To(BeTrue())
			Expect(cfg.PoolMaxIdleConns).To(Equal(10))
			Expect(cfg.PoolMaxOpenConns).To(Equal(100))
		})

		It("should handle config with GORM options", func() {
			cfg := &libgorm.Config{
				Driver:                 libgorm.DriverSQLite,
				DSN:                    ":memory:",
				SkipDefaultTransaction: true,
				PrepareStmt:            true,
				DryRun:                 false,
			}
			Expect(cfg.SkipDefaultTransaction).To(BeTrue())
			Expect(cfg.PrepareStmt).To(BeTrue())
			Expect(cfg.DryRun).To(BeFalse())
		})

		It("should handle disabled config", func() {
			cfg := &libgorm.Config{
				Driver:   libgorm.DriverSQLite,
				DSN:      ":memory:",
				Disabled: true,
			}
			Expect(cfg.Disabled).To(BeTrue())
		})
	})

	Describe("Validate", func() {
		It("should validate a correct config", func() {
			cfg := &libgorm.Config{
				Driver: libgorm.DriverSQLite,
				Name:   "test",
				DSN:    ":memory:",
			}
			err := cfg.Validate()
			Expect(err).To(BeNil())
		})

		It("should validate config without explicit validation tags", func() {
			// Config struct doesn't have validator tags, so validation passes
			cfg := &libgorm.Config{
				Name: "test",
				DSN:  ":memory:",
			}
			err := cfg.Validate()
			// Without validation tags, this should pass
			Expect(err).To(BeNil())
		})

		It("should validate minimal config", func() {
			cfg := &libgorm.Config{
				Driver: libgorm.DriverSQLite,
				DSN:    ":memory:",
			}
			err := cfg.Validate()
			Expect(err).To(BeNil())
		})

		It("should validate disabled config", func() {
			cfg := &libgorm.Config{
				Disabled: true,
			}
			err := cfg.Validate()
			Expect(err).To(BeNil())
		})
	})

	Describe("Config Methods", func() {
		It("should create GORM config from Config struct", func() {
			cfg := &libgorm.Config{
				Driver:                 libgorm.DriverSQLite,
				DSN:                    ":memory:",
				SkipDefaultTransaction: true,
				PrepareStmt:            true,
			}
			gormCfg := cfg.Config()
			Expect(gormCfg).ToNot(BeNil())
			Expect(gormCfg.SkipDefaultTransaction).To(BeTrue())
			Expect(gormCfg.PrepareStmt).To(BeTrue())
		})

		It("should register context function", func() {
			cfg := &libgorm.Config{
				Driver: libgorm.DriverSQLite,
				DSN:    ":memory:",
			}
			Expect(func() {
				cfg.RegisterContext(context.Background())
			}).ToNot(Panic())
		})
	})
})
