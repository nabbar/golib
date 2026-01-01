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

package semaphore_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsem "github.com/nabbar/golib/semaphore"
)

var _ = Describe("Progress Bar Creation", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(globalCtx, 5*time.Second)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("BarBytes", func() {
		It("should not create a bytes progress bar without MPB", func() {
			sem := libsem.New(ctx, 5, false)
			defer sem.DeferMain()

			bar := sem.BarBytes("Download", "file.zip", 1024*1024, false, nil)
			Expect(bar).ToNot(BeNil())
			Expect(bar.Total()).To(Equal(int64(0)))
		})

		It("should create a bytes progress bar with MPB", func() {
			sem := libsem.New(ctx, 5, true)
			defer sem.DeferMain()

			bar := sem.BarBytes("Download", "file.zip", 1024*1024, false, nil)
			Expect(bar).ToNot(BeNil())
			Expect(bar.Total()).To(Equal(int64(1024 * 1024)))

			// Simulate download
			bar.Inc64(512 * 1024)
			time.Sleep(10 * time.Millisecond)

			bar.Complete()
			time.Sleep(10 * time.Millisecond)
		})

		It("should queue bars sequentially", func() {
			sem := libsem.New(ctx, 5, true)
			defer sem.DeferMain()

			bar1 := sem.BarBytes("Download", "file1.zip", 1024, false, nil)
			bar2 := sem.BarBytes("Download", "file2.zip", 2048, false, bar1)

			Expect(bar1).ToNot(BeNil())
			Expect(bar2).ToNot(BeNil())

			bar1.Complete()
			bar2.Complete()
			time.Sleep(20 * time.Millisecond)
		})
	})

	Describe("BarTime", func() {
		It("should not create a time progress bar without MPB", func() {
			sem := libsem.New(ctx, 5, false)
			defer sem.DeferMain()

			bar := sem.BarTime("Process", "task", 100, false, nil)
			Expect(bar).ToNot(BeNil())
			Expect(bar.Total()).To(Equal(int64(0)))
		})

		It("should create a time progress bar with MPB", func() {
			sem := libsem.New(ctx, 5, true)
			defer sem.DeferMain()

			bar := sem.BarTime("Process", "task", 100, false, nil)
			Expect(bar).ToNot(BeNil())

			bar.Inc(50)
			time.Sleep(10 * time.Millisecond)

			bar.Complete()
			time.Sleep(10 * time.Millisecond)
		})
	})

	Describe("BarNumber", func() {
		It("should not create a number progress bar without MPB", func() {
			sem := libsem.New(ctx, 5, false)
			defer sem.DeferMain()

			bar := sem.BarNumber("Items", "processing", 1000, false, nil)
			Expect(bar).ToNot(BeNil())
			Expect(bar.Total()).To(Equal(int64(0)))
		})

		It("should create a number progress bar with MPB", func() {
			sem := libsem.New(ctx, 5, true)
			defer sem.DeferMain()

			bar := sem.BarNumber("Items", "processing", 1000, false, nil)
			Expect(bar).ToNot(BeNil())

			bar.Inc(100)
			time.Sleep(10 * time.Millisecond)

			bar.Complete()
			time.Sleep(10 * time.Millisecond)
		})
	})

	Describe("BarOpts", func() {
		It("should create a custom progress bar", func() {
			sem := libsem.New(ctx, 5, true)
			defer sem.DeferMain()

			bar := sem.BarOpts(500, false)
			Expect(bar).ToNot(BeNil())
			Expect(bar.Total()).To(Equal(int64(500)))

			bar.Complete()
			time.Sleep(10 * time.Millisecond)
		})

		It("should respect drop flag", func() {
			sem := libsem.New(ctx, 5, true)
			defer sem.DeferMain()

			bar := sem.BarOpts(100, true) // drop = true
			Expect(bar).ToNot(BeNil())

			bar.Inc(100)
			bar.Complete()
			time.Sleep(50 * time.Millisecond)

			Expect(bar.Completed()).To(BeTrue())
		})
	})

	Describe("Multiple progress bars", func() {
		It("should handle multiple concurrent bars", func() {
			sem := libsem.New(ctx, 5, true)
			defer sem.DeferMain()

			bars := make([]interface{}, 5)
			for i := 0; i < 5; i++ {
				bars[i] = sem.BarNumber("Task", "item", 100, false, nil)
				Expect(bars[i]).ToNot(BeNil())
			}

			time.Sleep(20 * time.Millisecond)
		})
	})
})
