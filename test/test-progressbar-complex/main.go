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
	"context"
	"fmt"
	"math/rand"
	"time"

	semtps "github.com/nabbar/golib/semaphore/types"

	libsem "github.com/nabbar/golib/semaphore"
	"github.com/vbauerster/mpb/v8"
)

func main() {
	tot := int64(100000)
	inc := int64(100)
	nbb := 5
	bar := make([]semtps.SemBar, 0)
	sem := libsem.New(context.Background(), 0, true, mpb.WithWidth(80), mpb.WithRefreshRate(10*time.Millisecond))

	defer func() {
		sem.DeferMain()
	}()

	println("\n\n\n")
	println("Starting complex...")

	for i := 0; i < nbb; i++ {
		if i > 0 && i%2 == 1 {
			bar = append(bar, sem.BarBytes("KiB bar", fmt.Sprintf("step %d", i), tot, false, bar[i-1]))
		} else if i > 0 && i%2 != 1 {
			bar = append(bar, sem.BarNumber("KiB bar", fmt.Sprintf("step %d", i), 0, false, bar[i-1]))
		} else {
			bar = append(bar, sem.BarTime("KiB bar", fmt.Sprintf("step %d", i), 0, false, nil))
		}
	}

	for i := range bar {
		if e := sem.NewWorker(); e != nil {
			panic(e)
		} else {
			go func(nb int, sem semtps.Sem) {
				defer func() {
					sem.DeferWorker()
				}()

				rand.Seed(999)

				for {
					if nb > 0 {
						if !bar[nb-1].Completed() {
							time.Sleep(time.Second)
							continue
						}
					}

					if bar[nb].Current() == 0 {
						bar[nb].Reset(tot, 0)
					}

					if bar[nb].Current() < tot {
						time.Sleep(getRandTime())
						bar[nb].Inc64(inc)
					} else {
						bar[nb].DeferMain()
						return
					}
				}
			}(i, sem)
		}
	}

	if e := sem.WaitAll(); e != nil {
		panic(e)
	}

	time.Sleep(500 * time.Millisecond)

	println("finish complex...")
}

func getRandTime() time.Duration {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	dur := time.Duration(rng.Intn(9)+1) * time.Millisecond

	if dur > 25*time.Millisecond {
		return 10 * time.Millisecond
	} else {
		return dur
	}
}
