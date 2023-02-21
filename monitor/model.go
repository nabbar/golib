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

package monitor

import (
	"context"
	"sync"

	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	montps "github.com/nabbar/golib/monitor/types"
)

const (
	defaultMonitorName = "not named"

	keyName        = "keyName"
	keyConfig      = "keyConfig"
	keyLogger      = "keyLogger"
	keyLoggerDef   = "keyLoggerDefault"
	keyHealthCheck = "keyFct"
	keyStatus      = "keySts"
	keyMessage     = "keyMsg"
	keyRise        = "keyRise"
	keyFall        = "keyFall"
	keyRun         = "keyRun"
	keyLastRun     = "keyLastRun"

	keyMetricsName = "keyMetricsName"
	keyMetricsFunc = "keyMetricsFunc"

	keyMetricLatency  = "metricLatency"
	keyMetricUpTime   = "metricUpTime"
	keyMetricDownTime = "metricDownTime"
	keyMetricRiseTime = "metricRiseTime"
	keyMetricFallTime = "metricFallTime"

	LogFieldProcess = "process"
	LogValueProcess = "monitor"
	LogFieldName    = "name"
)

type mon struct {
	m sync.RWMutex
	i montps.Info
	x libctx.Config[string]
	s chan struct{}
}

func (o *mon) SetHealthCheck(fct montps.HealthCheck) {
	o.x.Store(keyHealthCheck, fct)
}

func (o *mon) GetHealthCheck() montps.HealthCheck {
	return o.getFct()
}

func (o *mon) Clone(ctx context.Context) (montps.Monitor, liberr.Error) {
	n := &mon{}
	n.x = o.x.Clone(ctx)

	if o.IsRunning() {
		if e := n.Start(ctx); e != nil {
			if err, ok := e.(liberr.Error); ok {
				return nil, err
			} else {
				return nil, ErrorTimeout.ErrorParent(e)
			}
		}
	}

	return n, nil
}
