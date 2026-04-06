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

package errors

import (
	"fmt"
	"testing"
)

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New(1, "test error")
	}
}

func BenchmarkNewWithParent(b *testing.B) {
	parent := New(2, "parent error")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New(1, "test error", parent)
	}
}

func BenchmarkNewf(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Newf(1, "test error %s %d", "arg", 123)
	}
}

func BenchmarkAdd(b *testing.B) {
	err := New(1, "test error")
	parent := New(2, "parent error")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err.Add(parent)
	}
}

func BenchmarkIs(b *testing.B) {
	err1 := New(1, "test error")
	err2 := New(1, "test error")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err1.Is(err2)
	}
}

func BenchmarkHasCode(b *testing.B) {
	err := New(1, "test error", New(2, "parent error"))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.HasCode(CodeError(2))
	}
}

func BenchmarkError(b *testing.B) {
	err := New(1, "test error", New(2, "parent error"))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkGetTrace(b *testing.B) {
	err := New(1, "test error")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.GetTrace()
	}
}

func BenchmarkNewErrorTrace(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewErrorTrace(1, "test message", "/path/to/file.go", 10)
	}
}

func BenchmarkNewErrorRecovered(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewErrorRecovered("recovered message", "panic value")
	}
}

func BenchmarkFilterPath(b *testing.B) {
	path := "/sources/go/src/github.com/nabbar/golib/errors/some_file.go"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filterPath(path)
	}
}

func BenchmarkFilterPathVendor(b *testing.B) {
	path := "/sources/go/src/github.com/nabbar/golib/vendor/some/package/file.go"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filterPath(path)
	}
}

func BenchmarkFilterPathMod(b *testing.B) {
	path := "/go/pkg/mod/github.com/nabbar/golib/errors/some_file.go"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filterPath(path)
	}
}

func BenchmarkMake(b *testing.B) {
	stdErr := fmt.Errorf("standard error")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Make(stdErr)
	}
}

func BenchmarkMakeIfError(b *testing.B) {
	stdErr1 := fmt.Errorf("standard error 1")
	stdErr2 := fmt.Errorf("standard error 2")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MakeIfError(stdErr1, stdErr2)
	}
}

func BenchmarkAddOrNew(b *testing.B) {
	errMain := New(1, "main error")
	errSub := New(2, "sub error")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AddOrNew(errMain, errSub)
	}
}

func BenchmarkAddOrNewNilMain(b *testing.B) {
	errSub := New(2, "sub error")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AddOrNew(nil, errSub)
	}
}

func BenchmarkIfError(b *testing.B) {
	parent := New(1, "parent error")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IfError(1, "conditional error", parent)
	}
}

func BenchmarkIfErrorNoParent(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IfError(1, "conditional error")
	}
}
