/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package gorm_test

import (
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	logent "github.com/nabbar/golib/logger/entry"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
	"github.com/sirupsen/logrus"
	jww "github.com/spf13/jwalterweatherman"
)

// LogEntry represents a captured log entry for testing purposes.
// It stores all information passed to the logger during a log call.
type LogEntry struct {
	Level   loglvl.Level
	Message string
	Fields  map[string]interface{}
	Errors  []error
}

// MockEntry implements logent.Entry interface for testing.
// It captures log entry data and stores it in the parent MockLogger.
type MockEntry struct {
	level   loglvl.Level
	message string
	fields  map[string]interface{}
	errors  []error
	logger  *MockLogger
}

// NewMockEntry creates a new mock log entry with the specified logger and level.
func NewMockEntry(logger *MockLogger, level loglvl.Level) *MockEntry {
	return &MockEntry{
		level:  level,
		fields: make(map[string]interface{}),
		errors: make([]error, 0),
		logger: logger,
	}
}

// SetLogger sets the logger function for this entry (no-op in mock).
func (m *MockEntry) SetLogger(fct func() *logrus.Logger) logent.Entry {
	return m
}

// SetLevel sets the log level for this entry.
func (m *MockEntry) SetLevel(lvl loglvl.Level) logent.Entry {
	m.level = lvl
	return m
}

// SetMessageOnly configures message-only mode (no-op in mock).
func (m *MockEntry) SetMessageOnly(flag bool) logent.Entry {
	return m
}

// SetEntryContext sets entry context information including message.
func (m *MockEntry) SetEntryContext(etime time.Time, stack uint64, caller, file string, line uint64, msg string) logent.Entry {
	m.message = msg
	return m
}

// SetGinContext sets Gin context for this entry (no-op in mock).
func (m *MockEntry) SetGinContext(ctx *gin.Context) logent.Entry {
	return m
}

// DataSet sets arbitrary data for this entry (no-op in mock).
func (m *MockEntry) DataSet(data interface{}) logent.Entry {
	return m
}

// Check returns whether the entry contains errors.
func (m *MockEntry) Check(lvlNoErr loglvl.Level) bool {
	return len(m.errors) > 0
}

// Log finalizes the entry and captures it in the parent logger's entries list.
func (m *MockEntry) Log() {
	// Capture the log entry
	m.logger.entries = append(m.logger.entries, LogEntry{
		Level:   m.level,
		Message: m.message,
		Fields:  m.fields,
		Errors:  m.errors,
	})
}

// FieldAdd adds a structured field to this entry.
func (m *MockEntry) FieldAdd(key string, val interface{}) logent.Entry {
	m.fields[key] = val
	return m
}

// FieldMerge merges fields into this entry (no-op in mock).
func (m *MockEntry) FieldMerge(fields logfld.Fields) logent.Entry {
	return m
}

// FieldSet replaces all fields in this entry (no-op in mock).
func (m *MockEntry) FieldSet(fields logfld.Fields) logent.Entry {
	return m
}

// FieldClean removes specified fields from this entry.
func (m *MockEntry) FieldClean(keys ...string) logent.Entry {
	for _, key := range keys {
		delete(m.fields, key)
	}
	return m
}

// ErrorClean clears all errors from this entry.
func (m *MockEntry) ErrorClean() logent.Entry {
	m.errors = make([]error, 0)
	return m
}

// ErrorSet replaces all errors in this entry.
func (m *MockEntry) ErrorSet(err []error) logent.Entry {
	m.errors = err
	return m
}

// ErrorAdd appends errors to this entry, optionally filtering nil values.
func (m *MockEntry) ErrorAdd(cleanNil bool, err ...error) logent.Entry {
	for _, e := range err {
		if e != nil || !cleanNil {
			m.errors = append(m.errors, e)
		}
	}
	return m
}

// MockLogger implements liblog.Logger interface for testing.
// It captures log entries for assertion in tests.
type MockLogger struct {
	level   loglvl.Level
	entries []LogEntry
}

// NewMockLogger creates a new mock logger with default InfoLevel.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		level:   loglvl.InfoLevel,
		entries: make([]LogEntry, 0),
	}
}

// SetLevel sets the current log level for the mock logger.
func (m *MockLogger) SetLevel(lvl loglvl.Level) {
	m.level = lvl
}

// GetLevel returns the current log level of the mock logger.
func (m *MockLogger) GetLevel() loglvl.Level {
	return m.level
}

// Entry creates a new log entry with the specified level and message.
func (m *MockLogger) Entry(level loglvl.Level, message string, args ...interface{}) logent.Entry {
	entry := NewMockEntry(m, level)
	entry.message = message
	return entry
}

// Access creates an access log entry (no-op in mock).
func (m *MockLogger) Access(remoteAddr, remoteUser string, localtime time.Time, latency time.Duration, method, request, proto string, status int, size int64) logent.Entry {
	entry := NewMockEntry(m, loglvl.InfoLevel)
	return entry
}

func (m *MockLogger) SetOptions(opt *logcfg.Options) error {
	return nil
}

func (m *MockLogger) GetOptions() *logcfg.Options {
	return &logcfg.Options{}
}

func (m *MockLogger) Clone() (liblog.Logger, error) {
	return NewMockLogger(), nil
}

func (m *MockLogger) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *MockLogger) Close() error {
	return nil
}

func (m *MockLogger) SetIOWriterLevel(lvl loglvl.Level) {
	// No-op for mock
}

func (m *MockLogger) GetIOWriterLevel() loglvl.Level {
	return loglvl.InfoLevel
}

func (m *MockLogger) SetIOWriterFilter(pattern ...string) {
	// No-op for mock
}

func (m *MockLogger) AddIOWriterFilter(pattern ...string) {
	// No-op for mock
}

func (m *MockLogger) SetFields(field logfld.Fields) {
	// No-op for mock
}

func (m *MockLogger) GetFields() logfld.Fields {
	return logfld.New(nil)
}

func (m *MockLogger) SetSPF13Level(lvl loglvl.Level, log *jww.Notepad) {
	// No-op for mock
}

func (m *MockLogger) GetStdLogger(lvl loglvl.Level, logFlags int) *log.Logger {
	return log.New(io.Discard, "", 0)
}

func (m *MockLogger) SetStdLogger(lvl loglvl.Level, logFlags int) {
	// No-op for mock
}

func (m *MockLogger) Debug(message string, data interface{}, args ...interface{}) {
	// No-op for mock
}

func (m *MockLogger) Info(message string, data interface{}, args ...interface{}) {
	// No-op for mock
}

func (m *MockLogger) Warning(message string, data interface{}, args ...interface{}) {
	// No-op for mock
}

func (m *MockLogger) Error(message string, data interface{}, args ...interface{}) {
	// No-op for mock
}

func (m *MockLogger) Fatal(message string, data interface{}, args ...interface{}) {
	// No-op for mock
}

func (m *MockLogger) Panic(message string, data interface{}, args ...interface{}) {
	// No-op for mock
}

func (m *MockLogger) LogDetails(lvl loglvl.Level, message string, data interface{}, err []error, fields logfld.Fields, args ...interface{}) {
	// No-op for mock
}

func (m *MockLogger) CheckError(lvlKO, lvlOK loglvl.Level, message string, err ...error) bool {
	return false
}
