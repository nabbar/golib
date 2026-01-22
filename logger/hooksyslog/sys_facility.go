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

package hooksyslog

import "strings"

// Facility represents the facility code of a syslog message
// according to RFC 5424. The facility indicates the type of program
// or system component generating the message.
//
// Facilities are typically used for filtering and routing syslog messages:
//   - KERN: Kernel messages
//   - USER: User-level messages (default for applications)
//   - MAIL: Mail system
//   - DAEMON: System daemons
//   - AUTH: Security/authorization messages
//   - SYSLOG: Messages generated internally by syslogd
//   - LPR: Line printer subsystem
//   - NEWS: Network news subsystem
//   - UUCP: UUCP subsystem
//   - CRON: Clock daemon
//   - AUTHPRIV: Security/authorization messages (private)
//   - FTP: FTP daemon
//   - LOCAL0-LOCAL7: Reserved for local use (application-specific)
type Facility uint8

const (
	FacilityKern     Facility = iota // Kernel messages
	FacilityUser                     // User-level messages
	FacilityMail                     // Mail system
	FacilityDaemon                   // System daemons
	FacilityAuth                     // Security/authorization messages
	FacilitySyslog                   // Messages generated internally by syslogd
	FacilityLpr                      // Line printer subsystem
	FacilityNews                     // Network news subsystem
	FacilityUucp                     // UUCP subsystem
	FacilityCron                     // Clock daemon
	FacilityAuthPriv                 // Security/authorization messages (private)
	FacilityFTP                      // FTP daemon
	_                                // unused
	_                                // unused
	_                                // unused
	_                                // unused
	FacilityLocal0                   // Local use 0
	FacilityLocal1                   // Local use 1
	FacilityLocal2                   // Local use 2
	FacilityLocal3                   // Local use 3
	FacilityLocal4                   // Local use 4
	FacilityLocal5                   // Local use 5
	FacilityLocal6                   // Local use 6
	FacilityLocal7                   // Local use 7
)

// String returns the RFC 5424 name of the facility in uppercase.
// Returns an empty string for invalid/unknown facility values.
//
// Example:
//
//	fac := FacilityUser
//	fmt.Println(fac.String()) // Outputs: "USER"
func (f Facility) String() string {
	switch f {
	case FacilityKern:
		return "KERN"
	case FacilityUser:
		return "USER"
	case FacilityMail:
		return "MAIL"
	case FacilityDaemon:
		return "DAEMON"
	case FacilityAuth:
		return "AUTH"
	case FacilitySyslog:
		return "SYSLOG"
	case FacilityLpr:
		return "LPR"
	case FacilityNews:
		return "NEWS"
	case FacilityUucp:
		return "UUCP"
	case FacilityCron:
		return "CRON"
	case FacilityAuthPriv:
		return "AUTHPRIV"
	case FacilityFTP:
		return "FTP"
	case FacilityLocal0:
		return "LOCAL0"
	case FacilityLocal1:
		return "LOCAL1"
	case FacilityLocal2:
		return "LOCAL2"
	case FacilityLocal3:
		return "LOCAL3"
	case FacilityLocal4:
		return "LOCAL4"
	case FacilityLocal5:
		return "LOCAL5"
	case FacilityLocal6:
		return "LOCAL6"
	case FacilityLocal7:
		return "LOCAL7"
	}

	return ""
}

func (f Facility) Uint8() uint8 {
	return uint8(f)
}

// MakeFacility converts a facility string to a Facility value.
// The conversion is case-insensitive. Returns 0 if the string doesn't match any known facility.
func MakeFacility(facility string) Facility {
	switch strings.ToUpper(facility) {
	case FacilityKern.String():
		return FacilityKern
	case FacilityUser.String():
		return FacilityUser
	case FacilityMail.String():
		return FacilityMail
	case FacilityDaemon.String():
		return FacilityDaemon
	case FacilityAuth.String():
		return FacilityAuth
	case FacilitySyslog.String():
		return FacilitySyslog
	case FacilityLpr.String():
		return FacilityLpr
	case FacilityNews.String():
		return FacilityNews
	case FacilityUucp.String():
		return FacilityUucp
	case FacilityCron.String():
		return FacilityCron
	case FacilityAuthPriv.String():
		return FacilityAuthPriv
	case FacilityFTP.String():
		return FacilityFTP
	case FacilityLocal0.String():
		return FacilityLocal0
	case FacilityLocal1.String():
		return FacilityLocal1
	case FacilityLocal2.String():
		return FacilityLocal2
	case FacilityLocal3.String():
		return FacilityLocal3
	case FacilityLocal4.String():
		return FacilityLocal4
	case FacilityLocal5.String():
		return FacilityLocal5
	case FacilityLocal6.String():
		return FacilityLocal6
	case FacilityLocal7.String():
		return FacilityLocal7
	}

	return 0
}
