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

// Package pidcontroller_test contains the test suite for the pidcontroller package.
// It utilizes the Ginkgo BDD testing framework and Gomega matcher library to ensure
// the correctness and reliability of the PID controller implementation.
package pidcontroller_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
	Using https://onsi.github.io/ginkgo/
	Running with $> ginkgo -cover .
*/

// TestGolibPIDControler is the entry point for the PID Controller test suite.
// It registers the fail handler and runs the specifications defined in the suite.
// This function bridges the standard Go testing package with the Ginkgo framework.
func TestGolibPIDControler(t *testing.T) {
	// Registering fail handler to Gomega.
	// This connects Gomega's assertions (Expect) with Ginkgo's failure handling mechanism.
	RegisterFailHandler(Fail)

	// RunSpecs executes the test suite.
	// The description "PID Controler Suite" will appear in the test output.
	RunSpecs(t, "PID Controler Suite")
}

// BeforeSuite is run once before any of the specs in the suite are run.
// Use this to set up any global state or resources required for the tests.
var _ = BeforeSuite(func() {
	// Global initialization code goes here.
})

// AfterSuite is run once after all the specs in the suite have finished.
// Use this to clean up any global state or resources.
var _ = AfterSuite(func() {
	// Global cleanup code goes here.
})
