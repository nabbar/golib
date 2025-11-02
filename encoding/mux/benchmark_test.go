/*
MIT License

Copyright (c) 2023 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package mux_test

import (
	"bytes"
	"testing"

	encmux "github.com/nabbar/golib/encoding/mux"
)

func BenchmarkMultiplexerWrite(b *testing.B) {
	buf := &bytes.Buffer{}
	mux := encmux.NewMultiplexer(buf, '\n')
	channel := mux.NewChannel('a')
	data := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		channel.Write(data)
	}
}

func BenchmarkMultiplexerConcurrentWrite(b *testing.B) {
	buf := &bytes.Buffer{}
	mux := encmux.NewMultiplexer(buf, '\n')
	channel := mux.NewChannel('a')
	data := []byte("benchmark test data")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			channel.Write(data)
		}
	})
}

func BenchmarkDeMultiplexerRead(b *testing.B) {
	// Prepare multiplexed data
	buf := &bytes.Buffer{}
	mux := encmux.NewMultiplexer(buf, '\n')
	channel := mux.NewChannel('a')
	data := []byte("benchmark test data")

	for i := 0; i < 100; i++ {
		channel.Write(data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		bufCopy := bytes.NewBuffer(buf.Bytes())
		dmux := encmux.NewDeMultiplexer(bufCopy, '\n', 0)
		out := &bytes.Buffer{}
		dmux.NewChannel('a', out)
		b.StartTimer()

		dmux.Copy()
	}
}

func BenchmarkChannelRegistration(b *testing.B) {
	buf := &bytes.Buffer{}
	dmux := encmux.NewDeMultiplexer(buf, '\n', 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := &bytes.Buffer{}
		dmux.NewChannel(rune('a'+i%26), out)
	}
}

func BenchmarkChannelRegistrationConcurrent(b *testing.B) {
	buf := &bytes.Buffer{}
	dmux := encmux.NewDeMultiplexer(buf, '\n', 0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			out := &bytes.Buffer{}
			dmux.NewChannel(rune('a'+i%26), out)
			i++
		}
	})
}
