/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package hooksyslog

import (
	"sync/atomic"

	logcfg "github.com/nabbar/golib/logger/config"
	loglvl "github.com/nabbar/golib/logger/level"
	logtps "github.com/nabbar/golib/logger/types"
	libptc "github.com/nabbar/golib/network/protocol"
	"github.com/sirupsen/logrus"
)

type HookSyslog interface {
	logtps.Hook

	// Done returns a channel that will be closed when the hook is finished.
	// Use this channel to wait until the hook is finished and then flush
	// the buffer before exit function.
	//
	// It is important to note that this channel is not safe for concurrent
	// access from multiple goroutines. If you need to wait until the hook is
	// finished from multiple goroutines, you should use a wait group from the
	// sync package.
	//
	// Example:
	// h, _ := NewHookSyslog(LogOptionsSyslog{}, &logrus.JSONFormatter{})
	// go func() {
	// 	<-h.Done()
	// 	h.Flush()
	// }()
	//
	// It is important to note that you should not call Flush function after
	// getting the channel from Done function. This will block the hook from writing
	// new entries in the buffer.
	Done() <-chan struct{}
	// WriteSev writes a new entry in the syslog buffer with the specified severity
	// and data.
	//
	// The function returns the number of bytes written in the buffer and an
	// error if the write operation failed.
	//
	// The severity parameter should be one of the constants defined in the
	// SyslogSeverity type.
	//
	// The data parameter should contain the data to write in the syslog buffer.
	// It is important to note that the data parameter should not contain any
	// newline character, as it will be interpreted as the end of the syslog
	// entry.
	//
	// The hook will automatically add a newline character at the end of the
	// syslog entry if it is not already present.
	//
	// Example:
	// h, _ := NewHookSyslog(LogOptionsSyslog{}, &logrus.JSONFormatter{})
	// _, err = h.WriteSev(SyslogSeverityInfo, []byte("My syslog entry"))
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	//
	WriteSev(s SyslogSeverity, p []byte) (n int, err error)
}

// New returns a new HookSyslog instance.
//
// The opt parameter is required and cannot be nil.
//
// The function sets the syslog facility to the value of opt.Facility, the
// syslog severity to the value of opt.Severity, the network to the value of
// opt.Network, the host to the value of opt.Host, and the tag to the value of
// opt.Tag.
//
// The function then sets the levels of the hook to the value of opt.LogLevel.
// If the length of opt.LogLevel is less than 1, the function sets the levels to
// logrus.AllLevels.
//
// The function then creates a new io.Writer instance. If opt.DisableColor is true, it
// sets w to os.Stdout. Otherwise, it sets w to colorable.NewColorableStdout().
//
// The function then creates a new hkstd instance and sets its fields to w, lvls, f,
// opt.DisableStack, opt.DisableTimestamp, opt.EnableTrace, opt.DisableColor, and
// opt.EnableAccessLog.
//
// Finally, the function returns the new hkstd instance and no error.
func New(opt logcfg.OptionsSyslog, format logrus.Formatter) (HookSyslog, error) {
	var (
		LVLs = make([]logrus.Level, 0)
	)

	if len(opt.LogLevel) > 0 {
		for _, ls := range opt.LogLevel {
			LVLs = append(LVLs, loglvl.Parse(ls).Logrus())
		}
	} else {
		LVLs = logrus.AllLevels
	}

	n := &hks{
		s: new(atomic.Value),
		d: new(atomic.Value),
		o: ohks{
			format:           format,
			levels:           LVLs,
			disableStack:     opt.DisableStack,
			disableTimestamp: opt.DisableTimestamp,
			enableTrace:      opt.EnableTrace,
			enableAccessLog:  opt.EnableAccessLog,
			network:          libptc.Parse(opt.Network),
			endpoint:         opt.Host,
			tag:              opt.Tag,
			fac:              MakeFacility(opt.Facility),
			//sev : MakeSeverity(opt.Severity),
		},
		r: new(atomic.Bool),
	}

	n.s.Store(make(chan struct{}))
	n.d.Store(make(chan []data, 250))

	if h, e := n.getSyslog(); e != nil {
		return nil, e
	} else {
		_ = h.Close()
	}

	return n, nil
}
