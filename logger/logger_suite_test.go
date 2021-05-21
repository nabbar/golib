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

package logger_test

import (
	"context"
	"os"
	"testing"

	"github.com/nabbar/golib/ioutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	ctx context.Context
	cnl context.CancelFunc
)

/*
	Using https://onsi.github.io/ginkgo/
	Running with $> ginkgo -cover .
*/

func TestGolibAwsHelper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger Helper Suite")
}

var _ = BeforeSuite(func() {
	ctx, cnl = context.WithCancel(context.Background())
})

var _ = AfterSuite(func() {
	cnl()
})

func GetContext() context.Context {
	return ctx
}

func GetTempFile() (string, error) {
	hdf, err := ioutils.NewFileProgressTemp()
	if err != nil {
		return "", err
	}

	defer func() {
		_ = hdf.Close()
	}()

	return hdf.FilePath(), nil
}

func DelTempFile(filepath string) error {
	if _, err := os.Stat(filepath); err != nil {
		return err
	}
	return os.RemoveAll(filepath)
}
