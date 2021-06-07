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

import "strings"

type SyslogSeverity uint8

const (
	SyslogSeverityEmerg SyslogSeverity = iota + 1
	SyslogSeverityAlert
	SyslogSeverityCrit
	SyslogSeverityErr
	SyslogSeverityWarning
	SyslogSeverityNotice
	SyslogSeverityInfo
	SyslogSeverityDebug
)

func (s SyslogSeverity) String() string {
	switch s {
	case SyslogSeverityEmerg:
		return "EMERG"
	case SyslogSeverityAlert:
		return "ALERT"
	case SyslogSeverityCrit:
		return "CRIT"
	case SyslogSeverityErr:
		return "ERR"
	case SyslogSeverityWarning:
		return "WARNING"
	case SyslogSeverityNotice:
		return "NOTICE"
	case SyslogSeverityInfo:
		return "INFO"
	case SyslogSeverityDebug:
		return "DEBUG"
	}

	return ""
}

func MakeSeverity(severity string) SyslogSeverity {
	switch strings.ToUpper(severity) {
	case SyslogSeverityEmerg.String():
		return SyslogSeverityEmerg
	case SyslogSeverityAlert.String():
		return SyslogSeverityAlert
	case SyslogSeverityCrit.String():
		return SyslogSeverityCrit
	case SyslogSeverityErr.String():
		return SyslogSeverityErr
	case SyslogSeverityWarning.String():
		return SyslogSeverityWarning
	case SyslogSeverityNotice.String():
		return SyslogSeverityNotice
	case SyslogSeverityInfo.String():
		return SyslogSeverityInfo
	case SyslogSeverityDebug.String():
		return SyslogSeverityDebug
	}

	return 0
}

type SyslogFacility uint8

const (
	SyslogFacilityKern SyslogFacility = iota + 1
	SyslogFacilityUser
	SyslogFacilityMail
	SyslogFacilityDaemon
	SyslogFacilityAuth
	SyslogFacilitySyslog
	SyslogFacilityLpr
	SyslogFacilityNews
	SyslogFacilityUucp
	SyslogFacilityCron
	SyslogFacilityAuthPriv
	SyslogFacilityFTP
	SyslogFacilityLocal0
	SyslogFacilityLocal1
	SyslogFacilityLocal2
	SyslogFacilityLocal3
	SyslogFacilityLocal4
	SyslogFacilityLocal5
	SyslogFacilityLocal6
	SyslogFacilityLocal7
)

func (s SyslogFacility) String() string {
	switch s {
	case SyslogFacilityKern:
		return "KERN"
	case SyslogFacilityUser:
		return "USER"
	case SyslogFacilityMail:
		return "MAIL"
	case SyslogFacilityDaemon:
		return "DAEMON"
	case SyslogFacilityAuth:
		return "AUTH"
	case SyslogFacilitySyslog:
		return "SYSLOG"
	case SyslogFacilityLpr:
		return "LPR"
	case SyslogFacilityNews:
		return "NEWS"
	case SyslogFacilityUucp:
		return "UUCP"
	case SyslogFacilityCron:
		return "CRON"
	case SyslogFacilityAuthPriv:
		return "AUTHPRIV"
	case SyslogFacilityFTP:
		return "FTP"
	case SyslogFacilityLocal0:
		return "LOCAL0"
	case SyslogFacilityLocal1:
		return "LOCAL1"
	case SyslogFacilityLocal2:
		return "LOCAL2"
	case SyslogFacilityLocal3:
		return "LOCAL3"
	case SyslogFacilityLocal4:
		return "LOCAL4"
	case SyslogFacilityLocal5:
		return "LOCAL5"
	case SyslogFacilityLocal6:
		return "LOCAL6"
	case SyslogFacilityLocal7:
		return "LOCAL7"
	}

	return ""
}

func MakeFacility(facility string) SyslogFacility {
	switch strings.ToUpper(facility) {
	case SyslogFacilityKern.String():
		return SyslogFacilityKern
	case SyslogFacilityUser.String():
		return SyslogFacilityUser
	case SyslogFacilityMail.String():
		return SyslogFacilityMail
	case SyslogFacilityDaemon.String():
		return SyslogFacilityDaemon
	case SyslogFacilityAuth.String():
		return SyslogFacilityAuth
	case SyslogFacilitySyslog.String():
		return SyslogFacilitySyslog
	case SyslogFacilityLpr.String():
		return SyslogFacilityLpr
	case SyslogFacilityNews.String():
		return SyslogFacilityNews
	case SyslogFacilityUucp.String():
		return SyslogFacilityUucp
	case SyslogFacilityCron.String():
		return SyslogFacilityCron
	case SyslogFacilityAuthPriv.String():
		return SyslogFacilityAuthPriv
	case SyslogFacilityFTP.String():
		return SyslogFacilityFTP
	case SyslogFacilityLocal0.String():
		return SyslogFacilityLocal0
	case SyslogFacilityLocal1.String():
		return SyslogFacilityLocal1
	case SyslogFacilityLocal2.String():
		return SyslogFacilityLocal2
	case SyslogFacilityLocal3.String():
		return SyslogFacilityLocal3
	case SyslogFacilityLocal4.String():
		return SyslogFacilityLocal4
	case SyslogFacilityLocal5.String():
		return SyslogFacilityLocal5
	case SyslogFacilityLocal6.String():
		return SyslogFacilityLocal6
	case SyslogFacilityLocal7.String():
		return SyslogFacilityLocal7
	}

	return 0
}
