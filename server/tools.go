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

package server

import (
	"context"
	"time"
)

type FunCheck func() bool
type FunRun func()

func RunNbr(max uint8, chk FunCheck, run FunRun) bool {
	var i uint8

	for i = 0; i < max; i++ {
		if chk() {
			return true
		}

		run()
	}

	return chk()
}

func RunTick(ctx context.Context, tick, max time.Duration, chk FunCheck, run FunRun) bool {
	var (
		s = time.Now()
		t = time.NewTicker(tick)
	)

	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return false

		case <-t.C:
			if chk() {
				return true
			}

			run()

			if time.Since(s) >= max {
				return chk()
			}
		}
	}
}
