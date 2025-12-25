/*
 *  MIT License
 *
 *  Copyright (c) 2025 Nicolas JUHEL
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

package archive_test

import (
	"crypto/rand"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
	Using https://onsi.github.io/ginkgo/
	Running with $> ginkgo -cover .
*/

type EmptyStruct struct{}

var (
	err error
	lst = make(map[string]string, 0)
	arc = make(map[string]string, 0)
	dst string
)

func rnd(min, max int) int {
	diff := big.NewInt(int64(max - min))
	n, _ := rand.Int(rand.Reader, diff)
	return min + int(n.Int64())
}

// TestGolibEncodingAESHelper tests the Golib AES Encoding Helper function.
func TestGolibArchiveHelper(t *testing.T) {
	time.Sleep(500 * time.Millisecond)  // Adding delay for better testing synchronization
	RegisterFailHandler(Fail)           // Registering fail handler for better test failure reporting
	RunSpecs(t, "Archive Helper Suite") // Running the test suite for Encoding AES Helper
}

var _ = BeforeSuite(func() {
	for i := 1; i <= 5; i++ {
		p := filepath.Join(strings.Replace(reflect.TypeOf(EmptyStruct{}).PkgPath(), "_test", "", -1), "lorem_ipsum_"+strconv.Itoa(i)+".txt")
		lst[filepath.Base(p)] = p
	}

	dst = filepath.Join(strings.Replace(reflect.TypeOf(EmptyStruct{}).PkgPath(), "_test", "", -1), "extract_all_dir")

	var (
		w *os.File
		n int
		m = len(loremIpsum)
	)

	for f := range lst {
		w, err = os.Create(filepath.Base(f))
		Expect(err).ToNot(HaveOccurred())
		Expect(w).ToNot(BeNil())

		l := rnd(10, m)
		p := rnd(0, m-l)
		n, err = w.Write([]byte(loremIpsum[p : p+l]))

		Expect(err).ToNot(HaveOccurred())
		Expect(n).To(BeEquivalentTo(l))

		err = w.Sync()
		Expect(err).ToNot(HaveOccurred())

		err = w.Close()
		Expect(err).ToNot(HaveOccurred())
	}
})

var _ = AfterSuite(func() {
	for b := range lst {
		_ = os.Remove(b)
	}

	for _, b := range arc {
		_ = os.Remove(b)
	}

	for strings.Contains(dst, string(filepath.Separator)) && len(dst) > 3 {
		_ = os.RemoveAll(dst)
		dst = filepath.Dir(dst)
	}

	if len(dst) > 3 {
		_ = os.RemoveAll(dst)
	}
})
