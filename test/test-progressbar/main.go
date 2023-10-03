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
	tot := int64(1000)
	inc := int64(1)

	println("\n\n\n")
	println("Starting simple...")
	pb := progress.NewProgressBar(mpb.WithWidth(64), mpb.WithRefreshRate(200*time.Millisecond))

	brE := pb.NewBarSimpleETA("ETA bar", 0)
	brE.Reset(tot/2, 0)
	brE.Increment64(inc - 1)
	brE.Reset(tot, 0)

	brC := pb.NewBarSimpleCounter("counter bar", 0)
	brC.Reset(tot/2, 0)
	brC.Increment64(inc - 1)
	brC.Reset(tot, 0)

	brK := pb.NewBarSimpleKBits("KiB bar", 0)
	brK.Reset(tot/2, 0)
	brK.Increment64(inc - 1)
	brK.Reset(tot, 0)

	defer func() {
		brE.DeferMain()
		brC.DeferMain()
		brK.DeferMain()
	}()

	for i := int64(0); i < (tot / inc); i++ {
		time.Sleep(5 * time.Millisecond)

		if e := brE.NewWorker(); e != nil {
			println("Error : " + e.Error())
		} else {
			go func() {
				defer brE.DeferWorker()

				//nolint #nosec
				/* #nosec */
				rand.Seed(99)
				//nolint #nosec
				/* #nosec */
				time.Sleep(time.Duration(rand.Intn(9)) * time.Millisecond)

				brE.Increment64(inc - 1)
			}()
		}

		if e := brC.NewWorker(); e != nil {
			println("Error : " + e.Error())
		} else {
			go func() {
				defer brC.DeferWorker()

				//nolint #nosec
				/* #nosec */
				rand.Seed(99)
				//nolint #nosec
				/* #nosec */
				time.Sleep(time.Duration(rand.Intn(9)) * time.Millisecond)

				brC.Increment64(inc - 1)
			}()
		}

		if e := brK.NewWorker(); e != nil {
			println("Error : " + e.Error())
		} else {
			go func() {
				defer brK.DeferWorker()

				//nolint #nosec
				/* #nosec */
				rand.Seed(99)
				//nolint #nosec
				/* #nosec */
				time.Sleep(time.Duration(rand.Intn(9)) * time.Millisecond)

				brK.Increment64(inc - 1)
			}()
		}
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
