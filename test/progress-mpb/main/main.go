//go:build examples
// +build examples

/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package main

import (
	"fmt"
	"os"
	"time"

	tools "github.com/nabbar/golib/test/progress-mpb/tools"
)

func main() {
	pg := tools.MakeMPB(nil)

	b0 := tools.MakeBar(pg, nil, "bar 0")
	b1 := tools.MakeBar(pg, b0, "bar 1")
	b2 := tools.MakeBar(pg, b1, "bar 2")
	b3 := tools.MakeBar(pg, b2, "bar 3")
	b4 := tools.MakeBar(pg, b3, "bar 4")

	tot := int64(80)
	inc := int64(20)

	fail := func() {}
	log := func(msg string) {
		_, _ = fmt.Fprintln(os.Stderr, msg)
	}

	tools.Run(b0, time.Second, tot, inc, fail, log)
	tools.Run(b1, time.Second, tot, inc, fail, log)
	tools.Run(b2, time.Second, tot, inc, fail, log)
	tools.Run(b3, time.Second, tot, inc, fail, log)
	tools.Run(b4, time.Second, tot, inc, fail, log)

	pg.Wait()
}
