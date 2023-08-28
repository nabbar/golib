/*
 *  MIT License
 *
 *  Copyright (c) 2021 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package mailPooler

import (
	"context"
	"sync"
	"time"
)

type Counter interface {
	Pool(ctx context.Context) error
	Reset() error
	Clone() Counter
}

type counter struct {
	m   sync.Mutex
	num int

	max int
	dur time.Duration
	tim time.Time

	fct FuncCaller
}

func newCounter(max int, dur time.Duration, fct FuncCaller) Counter {
	return &counter{
		m:   sync.Mutex{},
		num: max,
		max: max,
		dur: dur,
		tim: time.Time{},
		fct: fct,
	}
}

func (c *counter) Pool(ctx context.Context) error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.max <= 0 || c.dur <= 0 {
		return nil
	}

	if e := ctx.Err(); e != nil {
		return ErrorMailPoolerContext.Error(e)
	}

	if c.tim.IsZero() {
		c.num = c.max
	} else if time.Since(c.tim) > c.dur {
		c.num = c.max
		c.tim = time.Time{}
	}

	if c.num > 0 {
		c.num--
		c.tim = time.Now()
	} else {
		time.Sleep(c.dur - time.Since(c.tim))

		c.num = c.max - 1
		c.tim = time.Now()

		if e := ctx.Err(); e != nil {
			return ErrorMailPoolerContext.Error(e)
		} else if err := c.fct(); err != nil {
			return err
		}
	}

	return nil
}

func (c *counter) Reset() error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.max <= 0 || c.dur <= 0 {
		return nil
	}

	c.num = c.max
	c.tim = time.Time{}

	if err := c.fct(); err != nil {
		return err
	}

	return nil
}

func (c *counter) Clone() Counter {
	return &counter{
		m:   sync.Mutex{},
		num: c.num,
		max: c.max,
		dur: c.dur,
		tim: time.Time{},
		fct: c.fct,
	}
}
