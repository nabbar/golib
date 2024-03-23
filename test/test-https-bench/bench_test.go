/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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
	"context"
	"fmt"
	"os"
	"testing"

	"golang.org/x/sys/cpu"
)

func BenchmarkRequest(b *testing.B) {
	var (
		ctx, cnl = context.WithCancel(context.Background())
		port     = GetFreePort()
		addr     = fmt.Sprintf(":%d", port)
	)

	defer cnl()

	_, _ = fmt.Fprintf(os.Stdout, "AVX: \t\t%v\n", cpu.X86.HasAVX)
	_, _ = fmt.Fprintf(os.Stdout, "AVX2: \t\t%v\n", cpu.X86.HasAVX2)
	_, _ = fmt.Fprintf(os.Stdout, "AVX512: \t%v\n", cpu.X86.HasAVX512)

	RunInit(ctx, addr)
	RunQuery(ctx, addr, b)
}
