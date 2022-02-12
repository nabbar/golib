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

package cobra

import (
	"fmt"
	"sort"

	liberr "github.com/nabbar/golib/errors"
	spfcbr "github.com/spf13/cobra"
)

func (c *cobra) AddCommandPrintErrorCode(fct FuncPrintErrorCode) {
	c.c.AddCommand(&spfcbr.Command{
		Use:     "error",
		Example: "error",
		Short:   "Print error code with package path related",
		Long:    "",
		Run: func(cmd *spfcbr.Command, args []string) {
			var (
				lst = liberr.GetCodePackages(c.getPackageGRootPath())
				key = make([]int, 0)
			)

			for c := range lst {
				key = append(key, int(c.GetUint16()))
			}

			sort.Ints(key)

			for _, c := range key {
				fct(fmt.Sprintf("%d", c), lst[liberr.CodeError(uint16(c))])
			}
		},
	})
}
