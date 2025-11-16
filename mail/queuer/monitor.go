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

package queuer

import (
	"context"
	"fmt"

	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

// Monitor creates a monitoring instance for health checking and observability.
//
// This method delegates to the underlying SMTP client's Monitor() method,
// providing access to the same monitoring capabilities as the raw SMTP client.
//
// The monitor can be used to:
//   - Perform periodic health checks
//   - Integrate with monitoring systems
//   - Track SMTP server availability
//   - Collect metrics and telemetry
//
// Parameters:
//   - ctx: Context for the monitor lifecycle
//   - vrs: Version information for monitoring metadata. See github.com/nabbar/golib/version
//     for creating version instances.
//
// Returns:
//   - montps.Monitor: A monitor instance for this SMTP connection
//   - error: ErrorParamEmpty-style error if SMTP client is not configured,
//     or any error from the underlying SMTP client's Monitor() method
//
// Note: The monitor operates on the underlying SMTP connection. Rate limiting
// does not apply to monitoring operations.
//
// Example:
//
//	version := libver.NewVersion(libver.License_MIT, "myapp", "MyApp",
//	    "2024-01-01", "prod", "1.0.0", "app", "", struct{}{}, 0)
//	monitor, err := pooler.Monitor(ctx, version)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Use monitor for health checks
//
// See also:
//   - github.com/nabbar/golib/monitor/types for Monitor interface details
//   - github.com/nabbar/golib/mail/smtp for SMTP monitoring capabilities
func (p *pooler) Monitor(ctx context.Context, vrs libver.Version) (montps.Monitor, error) {
	if p.s == nil {
		return nil, fmt.Errorf("SMTP client not defined")
	}

	return p.s.Monitor(ctx, vrs)
}
