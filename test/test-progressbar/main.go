//go:build examples
// +build examples

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
	max := int64(1000000)
	inc := int64(100)

	println("\n\n\n")
	println("Starting simple...")
	pb := progress.NewProgressBar(mpb.WithWidth(64), mpb.WithRefreshRate(200*time.Millisecond))

	brE := pb.NewBarSimpleETA("ETA bar", max)
	brE.Reset(max, 0)

	brC := pb.NewBarSimpleCounter("counter bar", max)
	brC.Reset(max, 0)

	brK := pb.NewBarSimpleKBits("KiB bar", max)
	brK.Reset(max, 0)

	defer func() {
		brE.DeferMain()
		brC.DeferMain()
		brK.DeferMain()
	}()

	for i := int64(0); i < (max / inc); i++ {
		time.Sleep(5 * time.Millisecond)

		if e := brE.NewWorker(); e != nil {
			println("Error : " + e.Error())
			continue
		}
		if e := brC.NewWorker(); e != nil {
			println("Error : " + e.Error())
			brE.DeferWorker()
			continue
		}
		if e := brK.NewWorker(); e != nil {
			println("Error : " + e.Error())
			brE.DeferWorker()
			brC.DeferWorker()
			continue
		}
		go func() {
			defer func() {
				brE.DeferWorker()
				brC.DeferWorker()
				brK.DeferWorker()
			}()

			//nolint #nosec
			/* #nosec */
			rand.Seed(99)
			//nolint #nosec
			/* #nosec */
			time.Sleep(time.Duration(rand.Intn(9)) * time.Millisecond)

			brE.Increment64(inc - 1)
			brC.Increment64(inc - 1)
			brK.Increment64(inc - 1)
		}()
	}

	if e := brE.WaitAll(); e != nil {
		panic(e)
	}
	if e := brC.WaitAll(); e != nil {
		panic(e)
	}
	if e := brK.WaitAll(); e != nil {
		panic(e)
	}

	brE.Done()
	brC.Done()
	brK.Done()

	time.Sleep(500 * time.Millisecond)

	println("finish simple...")
}
