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
	"fmt"
	"math/rand"
	"time"

	"github.com/nabbar/golib/progress"
	libsem "github.com/nabbar/golib/semaphore"
	"github.com/vbauerster/mpb/v5"
)

func main() {
	tot := int64(1000)
	inc := int64(1)
	nbb := 5
	bar := make([]progress.Bar, 0)

	println("\n\n\n")
	println("Starting complex...")

	pb := progress.NewProgressBar(mpb.WithWidth(64), mpb.WithRefreshRate(200*time.Millisecond))
	pb.MainProcessInit()

	defer func() {
		pb.DeferMain()
	}()

	for i := 0; i < nbb; i++ {
		if i > 0 && i%2 == 1 {
			bar = append(bar, pb.NewBarKBits("KiB bar", tot, fmt.Sprintf(" | step %d | ", i), bar[i-1]))
		} else if i > 0 && i%2 != 1 {
			bar = append(bar, pb.NewBarKBits("KiB bar", 0, fmt.Sprintf(" | step %d | ", i), bar[i-1]))
		} else {
			bar = append(bar, pb.NewBarKBits("KiB bar", 0, fmt.Sprintf(" | step %d | ", i), nil))
		}
	}

	for i := range bar {
		if e := pb.NewWorker(); e != nil {
			panic(e)
		} else {
			go func(nb int, sem libsem.Sem) {
				defer func() {
					pb.DeferWorker()
				}()

				rand.Seed(999)

				for {
					if nb > 0 && !bar[nb-1].Completed() {
						time.Sleep(time.Second)
					} else if bar[nb].Current() == 0 {
						bar[nb].Reset(tot, 0)
					}

					if bar[nb].Current() < tot {
						time.Sleep(time.Duration(rand.Intn(9)+1) * time.Millisecond)
						bar[nb].Increment64(inc)
					} else {
						bar[nb].Done()
						return
					}
				}
			}(i, pb)
		}
	}

	if e := pb.WaitAll(); e != nil {
		panic(e)
	}

	time.Sleep(500 * time.Millisecond)

	println("finish complex...")
}
