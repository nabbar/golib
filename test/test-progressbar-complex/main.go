/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
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

package main

import (
	"math/rand"
	"time"

	"github.com/nabbar/golib/progress"
	"github.com/vbauerster/mpb/v5"
)

func main() {
	max := int64(1000000000)
	inc := int64(100000)

	println("\n\n\n")
	println("Starting complex...")

	pb := progress.NewProgressBar(mpb.WithWidth(64), mpb.WithRefreshRate(200*time.Millisecond))
	pb.UnicProcessInit()

	defer func() {
		pb.UnicProcessDefer()
	}()

	brK0 := pb.NewBarKBits("KiB bar", 1, " | init | ", nil)
	brK1 := pb.NewBarKBits("KiB bar", max, " | step 1 | ", brK0)
	brK2 := pb.NewBarKBits("KiB bar", max, " | step 2 | ", brK1)

	brK0.Reset(max, 0)
	brK0.Done()

	var (
		done1 = false
		done2 = false
	)

	for i := int64(0); i < (max / inc); i++ {
		time.Sleep(1 * time.Millisecond)
		if e := pb.UnicProcessNewWorker(); e != nil {
			continue
		}
		go func() {
			defer func() {
				pb.UnicProcessDeferWorker()
			}()

			rand.Seed(999)

			for done1 {
				if !done1 && brK1.Completed() {
					done1 = true
					brK0.Done()
					brK1.Reset(max, 0)
				} else if !done1 {
					time.Sleep(time.Duration(rand.Intn(9)+1) * time.Millisecond)
				}
			}

			time.Sleep(time.Duration((rand.Intn(99)/3)+1) * time.Millisecond)
			brK1.Increment64(inc)
		}()
	}

	for i := int64(0); i < (max / inc); i++ {
		time.Sleep(1 * time.Millisecond)
		if e := pb.UnicProcessNewWorker(); e != nil {
			continue
		}
		go func() {
			defer func() {
				pb.UnicProcessDeferWorker()
			}()

			rand.Seed(999)

			for done2 {
				if !done2 && brK1.Completed() {
					done2 = true
					brK1.Done()
					brK2.Reset(max, 0)
				} else if !done2 {
					time.Sleep(time.Duration(rand.Intn(9)+1) * time.Millisecond)
				}
			}

			time.Sleep(time.Duration(rand.Intn(99)+1) * time.Millisecond)
			brK2.Increment64(inc)
		}()
	}

	if e := pb.UnicProcessWait(); e != nil {
		panic(e)
	}

	time.Sleep(500 * time.Millisecond)

	println("finish complex...")
}
