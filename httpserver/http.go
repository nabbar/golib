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
 */

package httpserver

import (
	"sync"
)

func ListenWaitNotify(allSrv ...HTTPServer) {
	var wg sync.WaitGroup
	wg.Add(len(allSrv))

	for _, s := range allSrv {
		go func(serv HTTPServer) {
			defer wg.Done()
			serv.Listen()
			serv.WaitNotify()
		}(s)
	}

	wg.Wait()
}

func Listen(allSrv ...HTTPServer) {
	for _, s := range allSrv {
		go func(serv HTTPServer) {
			serv.Listen()
		}(s)
	}
}

func Restart(allSrv ...HTTPServer) {
	for _, s := range allSrv {
		s.Restart()
	}
}

func Shutdown(allSrv ...HTTPServer) {
	for _, s := range allSrv {
		s.Shutdown()
	}
}

func IsRunning(allSrv ...HTTPServer) bool {
	for _, s := range allSrv {
		if s.IsRunning() {
			return true
		}
	}

	return false
}
