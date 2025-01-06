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

package certificates_test

import (
	"os"
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

const (
	keyFile = "test_ed25519.key"
	pubFile = "test_ed25519.pub"
)

// TestGolibEncodingAESHelper tests the Golib AES Encoding Helper function.
func TestGolibCertificatesHelper(t *testing.T) {
	time.Sleep(500 * time.Millisecond)       // Adding delay for better testing synchronization
	RegisterFailHandler(Fail)                // Registering fail handler for better test failure reporting
	RunSpecs(t, "Certificates Helper Suite") // Running the test suite for Encoding AES Helper
}

var _ = AfterSuite(func() {
	if _, e := os.Stat(keyFile); e == nil {
		Expect(os.Remove(keyFile)).ToNot(HaveOccurred())
	}
	if _, e := os.Stat(pubFile); e == nil {
		Expect(os.Remove(pubFile)).ToNot(HaveOccurred())
	}
})
