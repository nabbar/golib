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

package nobar_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsem "github.com/nabbar/golib/semaphore"
	semtps "github.com/nabbar/golib/semaphore/types"
)

var (
	// Global context for all tests
	globalCtx    context.Context
	globalCancel context.CancelFunc
)

func TestBar(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Semaphore Bar Suite")
}

var _ = BeforeSuite(func() {
	// Create a global context with timeout for all tests
	globalCtx, globalCancel = context.WithTimeout(context.Background(), 30*time.Second)
})

var _ = AfterSuite(func() {
	if globalCancel != nil {
		globalCancel()
	}
})

// Helper function to create a test semaphore without progress bar
func createTestSemaphore(ctx context.Context, nbrSimultaneous int) semtps.SemPgb {
	sem := libsem.New(ctx, nbrSimultaneous, false)
	Expect(sem).ToNot(BeNil())

	// Type assert to SemPgb
	semPgb, ok := sem.(semtps.SemPgb)
	Expect(ok).To(BeTrue(), "Semaphore should implement SemPgb")
	return semPgb
}

// Helper function to create a test semaphore WITH progress bar (MPB)
func createTestSemaphoreWithProgress(ctx context.Context, nbrSimultaneous int) semtps.SemPgb {
	sem := libsem.New(ctx, nbrSimultaneous, true)
	Expect(sem).ToNot(BeNil())

	// Type assert to SemPgb
	semPgb, ok := sem.(semtps.SemPgb)
	Expect(ok).To(BeTrue(), "Semaphore should implement SemPgb")
	return semPgb
}
