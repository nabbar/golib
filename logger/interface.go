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

package logger

import (
	"context"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"

	libctx "github.com/nabbar/golib/context"
	logcfg "github.com/nabbar/golib/logger/config"
	logent "github.com/nabbar/golib/logger/entry"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
	jww "github.com/spf13/jwalterweatherman"
)

// FuncLog is a function type that returns a Logger instance.
// This is used for dependency injection and lazy initialization of loggers.
type FuncLog func() Logger

// Logger is the main interface for structured logging operations.
// It extends io.WriteCloser to allow using the logger as a standard Go writer.
// The logger supports multiple output destinations (file, syslog, stdout/stderr),
// level-based filtering, custom fields injection, and integration with popular frameworks.
type Logger interface {
	io.WriteCloser

	//SetLevel allow to change the minimal level of log message
	SetLevel(lvl loglvl.Level)

	//GetLevel return the minimal level of log message
	GetLevel() loglvl.Level

	//SetIOWriterLevel allow to change the minimal level of log message for io.WriterCloser interface
	SetIOWriterLevel(lvl loglvl.Level)

	//GetIOWriterLevel return the minimal level of log message for io.WriterCloser interface
	GetIOWriterLevel() loglvl.Level

	// SetIOWriterFilter allow to filter message that contained the given pattern.
	// When received any message, if one pattern is found, the log is drop.
	// If the given pattern is empty, the pattern list is truncated to no one pattern.
	SetIOWriterFilter(pattern ...string)

	// SetIOWriterFilter allow to add a filter pattern into the filter message pattern list.
	AddIOWriterFilter(pattern ...string)

	//SetOptions allow to set or update the options for the logger
	SetOptions(opt *logcfg.Options) error

	//GetOptions return the options for the logger
	GetOptions() *logcfg.Options

	//SetFields allow to set or update the default fields for all logger entry
	// Fields are custom information added into log message
	SetFields(field logfld.Fields)

	//GetFields return the default fields for all logger entry
	// Fields are custom information added into log message
	GetFields() logfld.Fields

	//Clone allow to duplicate the logger with a copy of the logger
	Clone() (Logger, error)

	//SetSPF13Level allow to plus spf13 logger (jww) to this logger
	SetSPF13Level(lvl loglvl.Level, log *jww.Notepad)

	//GetStdLogger return a golang log.logger instance linked with this main logger.
	GetStdLogger(lvl loglvl.Level, logFlags int) *log.Logger

	//SetStdLogger force the default golang log.logger instance linked with this main logger.
	SetStdLogger(lvl loglvl.Level, logFlags int)

	//Debug add an entry with DebugLevel to the logger
	Debug(message string, data interface{}, args ...interface{})

	//Info add an entry with InfoLevel to the logger
	Info(message string, data interface{}, args ...interface{})

	//Warning add an entry with WarnLevel to the logger
	Warning(message string, data interface{}, args ...interface{})

	//Error add an entry with ErrorLevel level to the logger
	Error(message string, data interface{}, args ...interface{})

	//Fatal add an entry with FatalLevel to the logger
	//The function will break the process (os.exit) after log entry.
	Fatal(message string, data interface{}, args ...interface{})

	//Panic add an entry with PanicLevel level to the logger
	//The function will break the process (os.exit) after log entry.
	Panic(message string, data interface{}, args ...interface{})

	//LogDetails add an entry to the logger
	LogDetails(lvl loglvl.Level, message string, data interface{}, err []error, fields logfld.Fields, args ...interface{})

	//CheckError will check if a not nil error is given and if yes, will add an entry to the logger.
	// Othwise if the lvlOK is given (and not NilLevel) the function will add entry and said ok
	CheckError(lvlKO, lvlOK loglvl.Level, message string, err ...error) bool

	//Entry will return an entry struct to manage it (set gin context, add fields, log the entry...)
	Entry(lvl loglvl.Level, message string, args ...interface{}) logent.Entry

	//Access will return an entry struct to store info level access log message
	Access(remoteAddr, remoteUser string, localtime time.Time, latency time.Duration, method, request, proto string, status int, size int64) logent.Entry
}

// New returns a new Logger instance with the given context.
// The context is used to configure the logger fields.
// The logger level is set to InfoLevel by default.
// The returned logger can be used to log messages at different levels.
// The logger can also be used to set the default golang log.logger instance.
func New(ctx context.Context) Logger {
	l := &logger{
		m: sync.RWMutex{},
		x: libctx.New[uint8](ctx),
		f: logfld.New(ctx),
		c: new(atomic.Value),
	}

	l.SetLevel(loglvl.InfoLevel)

	return l
}
