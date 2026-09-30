/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package opengpg_test

import (
	"bytes"
	"context"
	"runtime"
	"testing"

	libgpg "github.com/nabbar/golib/encoding/opengpg"
)

func doReadTest(b *testing.B, mod libgpg.OpenGPG) {
	var (
		err error
		idt = &libgpg.Identity{
			Name:    "Bench User",
			Comment: "Benchmark",
			Email:   "bench@example.com",
		}
		buf = bytes.NewBuffer(make([]byte, 0, 10*tstDataSize))
		res = bytes.NewBuffer(make([]byte, 0, tstDataSize))
	)

	b.Cleanup(func() {
		if mod != nil {
			_ = mod.Close()
		}
		runtime.GC()
	})

	if err = mod.Create(idt); err != nil {
		b.Fatalf("failed to setup key: %v", err)
	}

	b.ReportAllocs()
	runtime.GC()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		readerOpenGPG(b, mod, buf, res, []byte(tstDataBase))
	}

	b.StopTimer()

	if dur := b.Elapsed().Seconds(); dur > 0 {
		b.ReportMetric(float64(b.N)/dur, "ops/s")
	}
}

func doWriteTest(b *testing.B, mod libgpg.OpenGPG) {
	var (
		err error
		idt = &libgpg.Identity{
			Name:    "Bench User",
			Comment: "Benchmark",
			Email:   "bench@example.com",
		}
		buf = bytes.NewBuffer(make([]byte, 0, 10*tstDataSize))
		res = bytes.NewBuffer(make([]byte, 0, tstDataSize))
	)

	b.Cleanup(func() {
		if mod != nil {
			_ = mod.Close()
		}
		runtime.GC()
	})

	if err = mod.Create(idt); err != nil {
		b.Fatalf("failed to setup key: %v", err)
	}

	b.ReportAllocs()
	runtime.GC()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		writerOpenGPG(b, mod, buf, res, []byte(tstDataBase))
	}

	b.StopTimer()

	if dur := b.Elapsed().Seconds(); dur > 0 {
		b.ReportMetric(float64(b.N)/dur, "ops/s")
	}
}

func BenchmarkReadMLDSA65ED(b *testing.B) {
	doReadTest(b, getNewOpenGPGPMLDSA65(context.Background()))
}

func BenchmarkReadMLDSA87ED(b *testing.B) {
	doReadTest(b, getNewOpenGPGPMLDSA87(context.Background()))
}

func BenchmarkReadSLHDSA128f(b *testing.B) {
	doReadTest(b, getNewOpenGPGPSLHDSA128f(context.Background()))
}

func BenchmarkReadSLHDSA128s(b *testing.B) {
	doReadTest(b, getNewOpenGPGPSLHDSA128s(context.Background()))
}

func BenchmarkReadSLHDSA256s(b *testing.B) {
	doReadTest(b, getNewOpenGPGPSLHDSA256s(context.Background()))
}

func BenchmarkWriteMLDSA65ED(b *testing.B) {
	doWriteTest(b, getNewOpenGPGPMLDSA65(context.Background()))
}

func BenchmarkWriteMLDSA87ED(b *testing.B) {
	doWriteTest(b, getNewOpenGPGPMLDSA87(context.Background()))
}

func BenchmarkWriteSLHDSA128f(b *testing.B) {
	doWriteTest(b, getNewOpenGPGPSLHDSA128f(context.Background()))
}

func BenchmarkWriteSLHDSA128s(b *testing.B) {
	doWriteTest(b, getNewOpenGPGPSLHDSA128s(context.Background()))
}

func BenchmarkWriteSLHDSA256s(b *testing.B) {
	doWriteTest(b, getNewOpenGPGPSLHDSA256s(context.Background()))
}
