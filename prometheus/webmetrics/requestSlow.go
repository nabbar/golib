/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
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

package webmetrics

import (
	"context"
	"fmt"
	"strconv"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	libprm "github.com/nabbar/golib/prometheus"
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
	librtr "github.com/nabbar/golib/router"
)

func MetricRequestSlow(prefixName string, fct libprm.FuncGetPrometheus) prmmet.Metric {
	var (
		met prmmet.Metric
		prm libprm.Prometheus
	)

	if fct == nil {
		return nil
	} else if prm = fct(); prm == nil {
		return nil
	}

	met = prmmet.NewMetrics(getDefaultPrefix(prefixName, "request_slow_total"), prmtps.Histogram)
	met.SetDesc(fmt.Sprintf("the server handled slow requests counter, t=%d.", prm.GetSlowTime()))
	met.AddLabel("uri", "method", "code")
	met.AddBuckets(prm.GetDuration()...)
	met.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		var (
			c  *ginsdk.Context
			ok bool
		)

		if c, ok = ctx.(*ginsdk.Context); !ok {
			return
		} else if fct == nil {
			return
		} else if prm = fct(); prm == nil {
			return
		} else if start := c.GetInt64(librtr.GinContextStartUnixNanoTime); start == 0 {
			return
		} else if ts := time.Unix(0, start); ts.IsZero() {
			return
		} else if int32(time.Since(ts).Seconds()) > prm.GetSlowTime() {
			_ = m.Inc([]string{c.FullPath(), c.Request.Method, strconv.Itoa(c.Writer.Status())})
		}
	})

	return met
}
