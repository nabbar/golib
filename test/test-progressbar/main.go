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

	njs_progress "github.com/nabbar/golib/njs-progress"
	"github.com/vbauerster/mpb/v5"
)

var (
	pb njs_progress.ProgressBar
	br njs_progress.Bar
)

func main() {
	println("Starting...")

	pb = njs_progress.NewProgressBar(0, time.Time{}, nil, mpb.WithWidth(64))
	pb.SetSemaphoreOption(0, 0)
	br = pb.NewBarSimple("test bar")

	defer br.DeferMain(false)

	for i := 0; i < 1000; i++ {

		if e := br.NewWorker(); e != nil {
			println("Error : " + e.Error())
			continue
		}

		go func(id int) {
			defer br.DeferWorker()
			rand.Seed(9999)
			time.Sleep(time.Duration(rand.Intn(999)) * time.Millisecond)
		}(i)
	}

	if e := br.WaitAll(); e != nil {
		panic(e)
	}

	println("finish...")
}
