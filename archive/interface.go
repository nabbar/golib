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

package archive

import (
	"io"

	arcarc "github.com/nabbar/golib/archive/archive"
	arctps "github.com/nabbar/golib/archive/archive/types"
	arccmp "github.com/nabbar/golib/archive/compress"
)

func ParseCompression(s string) arccmp.Algorithm {
	return arccmp.Parse(s)
}

func DetectCompression(r io.Reader) (arccmp.Algorithm, io.ReadCloser, error) {
	return arccmp.Detect(r)
}

func ParseArchive(s string) arcarc.Algorithm {
	return arcarc.Parse(s)
}

func DetectArchive(r io.ReadCloser) (arcarc.Algorithm, arctps.Reader, io.ReadCloser, error) {
	return arcarc.Detect(r)
}
