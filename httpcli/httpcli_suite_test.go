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

package httpcli_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
	Using https://onsi.github.io/ginkgo/
	Running with $> ginkgo -cover .
*/

var srv = &http.Server{
	Addr:    ":8080",
	Handler: Hello(),
}

var (
	ctx context.Context
	cnl context.CancelFunc
)

func TestGolibHttpCliHelper(t *testing.T) {
	defer func() {
		_ = srv.Shutdown(context.Background())
	}()

	ctx, cnl = context.WithCancel(context.Background())
	defer cnl()

	go func() {
		if e := srv.ListenAndServe(); e != nil {
			if !errors.Is(e, http.ErrServerClosed) {
				panic(e)
			}
		}
	}()

	time.Sleep(500 * time.Millisecond)

	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Cli Helper Suite")
}

var _ = BeforeSuite(func() {
})

var _ = AfterSuite(func() {
})

func Hello() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintf(writer, "hello\n")
		_, _ = fmt.Fprintf(writer, "Requested Hostname: %s\n", request.Host)
		_, _ = fmt.Fprintf(writer, "Requested Uri: %s\n", request.RequestURI)
		_, _ = fmt.Fprintf(writer, "Requested: %s\n", request.RemoteAddr)
	}
}
