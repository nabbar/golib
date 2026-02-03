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
	"io"

	loglvl "github.com/nabbar/golib/logger/level"
	jww "github.com/spf13/jwalterweatherman"
)

// SetSPF13Level configures the global jwalterweatherman logger (used by Hugo, Cobra, Viper)
// to use this logger as its output destination.
//
// jwalterweatherman is the logging library used by spf13 projects like Hugo and Cobra.
// This method bridges those projects' logging to this logger.
//
// The method maps log levels as follows:
//   - NilLevel: Disables jww logging (output to io.Discard)
//   - DebugLevel: jww.LevelTrace (most verbose)
//   - InfoLevel: jww.LevelInfo
//   - WarnLevel: jww.LevelWarn
//   - ErrorLevel: jww.LevelError
//   - FatalLevel: jww.LevelFatal
//   - PanicLevel: jww.LevelCritical (most severe)
//
// Parameters:
//   - lvl: The minimum log level for jww messages
//   - log: The jww.Notepad instance (pass nil to disable jww stdout)
//
// Example:
//
//	// Capture Hugo/Cobra logs
//	logger.SetSPF13Level(loglvl.InfoLevel, nil)
//	// Now all jww.INFO.Println() calls will use this logger
func (o *lgr) SetSPF13Level(lvl loglvl.Level, log *jww.Notepad) {
	if log == nil {
		jww.SetStdoutOutput(io.Discard)
	} else {
		jww.SetStdoutOutput(o)
	}

	switch lvl {
	case loglvl.NilLevel:
		jww.SetLogOutput(io.Discard)
		jww.SetLogThreshold(jww.LevelCritical)
	case loglvl.DebugLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelTrace)
	case loglvl.InfoLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelInfo)
	case loglvl.WarnLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelWarn)
	case loglvl.ErrorLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelError)
	case loglvl.FatalLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelFatal)
	case loglvl.PanicLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelCritical)
	}
}
