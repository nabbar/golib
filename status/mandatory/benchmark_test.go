/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package mandatory_test

import (
	"testing"

	stsctr "github.com/nabbar/golib/status/control"
	"github.com/nabbar/golib/status/mandatory"
)

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mandatory.New()
	}
}

func BenchmarkSetMode(b *testing.B) {
	m := mandatory.New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.SetMode(stsctr.Should)
	}
}

func BenchmarkGetMode(b *testing.B) {
	m := mandatory.New()
	m.SetMode(stsctr.Should)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.GetMode()
	}
}

func BenchmarkKeyAdd(b *testing.B) {
	m := mandatory.New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.KeyAdd("key1")
	}
}

func BenchmarkKeyAddMultiple(b *testing.B) {
	m := mandatory.New()
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.KeyAdd(keys...)
	}
}

func BenchmarkKeyHas(b *testing.B) {
	m := mandatory.New()
	m.KeyAdd("key1", "key2", "key3")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.KeyHas("key2")
	}
}

func BenchmarkKeyDel(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		m := mandatory.New()
		m.KeyAdd("key1", "key2", "key3")
		b.StartTimer()
		m.KeyDel("key2")
	}
}

func BenchmarkKeyList(b *testing.B) {
	m := mandatory.New()
	m.KeyAdd("key1", "key2", "key3", "key4", "key5")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.KeyList()
	}
}

func BenchmarkKeyListLarge(b *testing.B) {
	m := mandatory.New()
	for i := 0; i < 1000; i++ {
		m.KeyAdd(string(rune(i)))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.KeyList()
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	m := mandatory.New()
	m.KeyAdd("key1", "key2", "key3")
	m.SetMode(stsctr.Should)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.GetMode()
			_ = m.KeyHas("key1")
			_ = m.KeyList()
		}
	})
}

func BenchmarkConcurrentWrites(b *testing.B) {
	m := mandatory.New()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.SetMode(stsctr.Should)
			m.KeyAdd("key1")
		}
	})
}

func BenchmarkMixedOperations(b *testing.B) {
	m := mandatory.New()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 6 {
			case 0:
				m.SetMode(stsctr.Should)
			case 1:
				_ = m.GetMode()
			case 2:
				m.KeyAdd("key1")
			case 3:
				_ = m.KeyHas("key1")
			case 4:
				m.KeyDel("key1")
			case 5:
				_ = m.KeyList()
			}
			i++
		}
	})
}
