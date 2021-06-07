// +build !windows

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

package logger

import (
	"fmt"
	"log/syslog"
)

func makePriority(severity SyslogSeverity, facility SyslogFacility) syslog.Priority {
	return makePriorotySeverity(severity) | makePriorotyFacility(facility)
}

func makePriorotySeverity(sev SyslogSeverity) syslog.Priority {
	switch sev {
	case SyslogSeverityEmerg:
		return syslog.LOG_EMERG
	case SyslogSeverityAlert:
		return syslog.LOG_ALERT
	case SyslogSeverityCrit:
		return syslog.LOG_CRIT
	case SyslogSeverityErr:
		return syslog.LOG_ERR
	case SyslogSeverityWarning:
		return syslog.LOG_WARNING
	case SyslogSeverityNotice:
		return syslog.LOG_NOTICE
	case SyslogSeverityInfo:
		return syslog.LOG_INFO
	case SyslogSeverityDebug:
		return syslog.LOG_DEBUG
	}
	return 0
}

func makePriorotyFacility(fac SyslogFacility) syslog.Priority {
	switch fac {
	case SyslogFacilityKern:
		return syslog.LOG_KERN
	case SyslogFacilityUser:
		return syslog.LOG_USER
	case SyslogFacilityMail:
		return syslog.LOG_MAIL
	case SyslogFacilityDaemon:
		return syslog.LOG_DAEMON
	case SyslogFacilityAuth:
		return syslog.LOG_AUTH
	case SyslogFacilitySyslog:
		return syslog.LOG_SYSLOG
	case SyslogFacilityLpr:
		return syslog.LOG_LPR
	case SyslogFacilityNews:
		return syslog.LOG_NEWS
	case SyslogFacilityUucp:
		return syslog.LOG_UUCP
	case SyslogFacilityCron:
		return syslog.LOG_CRON
	case SyslogFacilityAuthPriv:
		return syslog.LOG_AUTHPRIV
	case SyslogFacilityFTP:
		return syslog.LOG_FTP
	case SyslogFacilityLocal0:
		return syslog.LOG_LOCAL0
	case SyslogFacilityLocal1:
		return syslog.LOG_LOCAL1
	case SyslogFacilityLocal2:
		return syslog.LOG_LOCAL2
	case SyslogFacilityLocal3:
		return syslog.LOG_LOCAL3
	case SyslogFacilityLocal4:
		return syslog.LOG_LOCAL4
	case SyslogFacilityLocal5:
		return syslog.LOG_LOCAL5
	case SyslogFacilityLocal6:
		return syslog.LOG_LOCAL6
	case SyslogFacilityLocal7:
		return syslog.LOG_LOCAL7
	}
	return 0
}

type _Syslog struct {
	w *syslog.Writer
}

func newSyslog(net NetworkType, host, tag string, severity SyslogSeverity, facility SyslogFacility) (syslogWrapper, error) {
	var (
		sys *syslog.Writer
		err error
	)

	if sys, err = syslog.Dial(net.String(), host, makePriority(severity, facility), tag); err != nil {
		return nil, err
	}

	return &_Syslog{
		w: sys,
	}, nil
}

func (o *_Syslog) Write(p []byte) (n int, err error) {
	if o.w == nil {
		return 0, fmt.Errorf("logrus.hooksyslog: connection not setup")
	}

	return o.w.Write(p)
}

func (o *_Syslog) Close() error {
	err := o.w.Close()
	o.w = nil
	return err
}

func (o *_Syslog) Panic(p []byte) (n int, err error) {
	return o.Write(p)
}

func (o *_Syslog) Fatal(p []byte) (n int, err error) {
	return o.Write(p)
}

func (o *_Syslog) Error(p []byte) (n int, err error) {
	return o.Write(p)
}

func (o *_Syslog) Warning(p []byte) (n int, err error) {
	return o.Write(p)
}

func (o *_Syslog) Info(p []byte) (n int, err error) {
	return o.Write(p)
}

func (o *_Syslog) Debug(p []byte) (n int, err error) {
	return o.Write(p)
}
