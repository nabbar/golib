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

package smtp

import (
	"context"
	"fmt"
	"runtime"

	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

const (
	// defaultNameMonitor is the default display name for the SMTP monitor.
	// This appears in monitoring dashboards and logs.
	defaultNameMonitor = "SMTP Client"
)

// HealthCheck performs a health check on the SMTP connection.
// This method is designed to be called by monitoring systems and provides
// a simple boolean health status based on the Check method.
//
// The health check verifies that the SMTP server is reachable and responding
// by establishing a connection and sending a NOOP command. The connection is
// closed after the check.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control (recommended: 5-10 seconds)
//
// Returns:
//   - error: nil if healthy; an error describing the issue otherwise
//
// This is the implementation of the health check function interface used by
// github.com/nabbar/golib/monitor/types.Monitor.
func (s *smtpClient) HealthCheck(ctx context.Context) error {
	return s.Check(ctx)
}

// Monitor creates and starts a monitoring instance for this SMTP client.
// The monitor provides:
//   - Periodic health checks
//   - Runtime information (Go version, build info)
//   - Server connection details (host, port)
//   - Integration with monitoring systems
//
// The monitor is automatically started and will begin performing health checks
// according to its configured interval. The caller should stop the monitor
// when no longer needed using monitor.Stop(ctx).
//
// Parameters:
//   - ctx: Context for the monitor's lifecycle
//   - vrs: Version information for metadata (release, build, date)
//
// Returns:
//   - montps.Monitor: A started monitor instance
//   - error: Any error during monitor creation or startup
//
// Example:
//
//	monitor, err := client.Monitor(ctx, version)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer monitor.Stop(ctx)
//
// See github.com/nabbar/golib/monitor/types for more information about
// monitor capabilities and configuration.
func (s *smtpClient) Monitor(ctx context.Context, vrs libver.Version) (montps.Monitor, error) {
	var (
		e   error
		inf moninf.Info
		mon montps.Monitor
		res = make(map[string]interface{}, 0)
	)

	res["runtime"] = runtime.Version()[2:]
	res["release"] = vrs.GetRelease()
	res["build"] = vrs.GetBuild()
	res["date"] = vrs.GetDate()

	if inf, e = moninf.New(defaultNameMonitor); e != nil {
		return nil, e
	} else {
		inf.RegisterName(func() (string, error) {
			return fmt.Sprintf("%s [%s:%d]", defaultNameMonitor, s.cfg.GetHost(), s.cfg.GetPort()), nil
		})
		inf.RegisterInfo(func() (map[string]interface{}, error) {
			return res, nil
		})
	}

	if mon, e = libmon.New(ctx, inf); e != nil {
		return nil, e
	}

	mon.SetHealthCheck(s.HealthCheck)
	if e = mon.Start(ctx); e != nil {
		return nil, e
	}

	return mon, nil
}
