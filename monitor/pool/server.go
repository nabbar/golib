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

package pool

import (
	"context"
	"fmt"
	"strings"
	"time"

	montps "github.com/nabbar/golib/monitor/types"
)

// Start starts all monitors in the pool.
// Returns an error if any monitor fails to start.
func (o *pool) Start(ctx context.Context) error {
	var err = make([]string, 0)
	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if e := val.Start(ctx); e != nil {
			err = append(err, fmt.Sprintf("error on starting monitor '%s': %v", val.Name(), e))
		} else if e = o.MonitorSet(val); e != nil {
			err = append(err, fmt.Sprintf("error on starting monitor '%s': %v", val.Name(), e))
		}
		return true
	})

	if len(err) > 0 {
		return fmt.Errorf("%s", strings.Join(err, ", \n"))
	}

	return nil
}

// Stop stops all monitors in the pool.
// Returns an error if any monitor fails to stop.
func (o *pool) Stop(ctx context.Context) error {
	var err = make([]string, 0)
	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if e := val.Stop(ctx); e != nil {
			err = append(err, fmt.Sprintf("error on stopping monitor '%s': %v", val.Name(), e))
		} else if e = o.MonitorSet(val); e != nil {
			err = append(err, fmt.Sprintf("error on stopping monitor '%s': %v", val.Name(), e))
		}
		return true
	})

	if len(err) > 0 {
		return fmt.Errorf("%s", strings.Join(err, ", \n"))
	}

	return nil
}

// Restart restarts all monitors in the pool.
// Returns an error if any monitor fails to restart.
func (o *pool) Restart(ctx context.Context) error {
	var err = make([]string, 0)
	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if e := val.Restart(ctx); e != nil {
			err = append(err, fmt.Sprintf("error on restarting monitor '%s': %v", val.Name(), e))
		} else if e = o.MonitorSet(val); e != nil {
			err = append(err, fmt.Sprintf("error on restarting monitor '%s': %v", val.Name(), e))
		}
		return true
	})

	if len(err) > 0 {
		return fmt.Errorf("%s", strings.Join(err, ", \n"))
	}

	return nil
}

// IsRunning returns true if at least one monitor in the pool is running.
func (o *pool) IsRunning() bool {
	var res = false

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		res = val.IsRunning()
		return !res
	})

	return res
}

// Uptime returns the maximum uptime among all monitors in the pool.
func (o *pool) Uptime() time.Duration {
	var res time.Duration

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if dur := val.Uptime(); res < dur {
			res = dur
		}

		return true
	})

	return res

}
