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

package njs_console

import (
	"bufio"
	"fmt"

	"github.com/fatih/color"
)

type colorType uint8

var (
	colorList map[colorType]*color.Color
)

const (
	ColorPrint colorType = iota
	ColorPrompt
)

func init() {
	colorList = map[colorType]*color.Color{
		ColorPrint:  nil,
		ColorPrompt: nil,
	}
}

func GetColorType(colId uint8) colorType {
	return colorType(colId)
}

func SetColor(col colorType, value ...int) {
	var cols = make([]color.Attribute, 0)

	for _, v := range value {
		cols = append(cols, color.Attribute(v))
	}

	colorList[col] = color.New(cols...)
}

func (c colorType) SetColor(col *color.Color) {
	colorList[c] = col
}

func (c colorType) Println(text string) {
	if colorList[c] != nil {
		_, _ = colorList[c].Println(text) // #nosec
	} else {
		println(text)
	}
}

func (c colorType) Print(text string) {
	if colorList[c] != nil {
		_, _ = colorList[c].Print(text) // #nosec
	} else {
		print(text)
	}
}

func (c colorType) BuffPrintf(buff *bufio.ReadWriter, format string, args ...interface{}) (n int, err error) {
	if colorList[c] != nil && buff != nil {
		return colorList[c].Fprintf(buff, format, args...) // #nosec
	} else if buff != nil {
		return buff.Write([]byte(fmt.Sprintf(format, args...)))
	} else {
		return 0, fmt.Errorf("buffer is not defined")
	}
}

func (c colorType) Sprintf(format string, args ...interface{}) string {
	if colorList[c] != nil {
		return colorList[c].Sprintf(format, args...) // #nosec
	} else {
		return fmt.Sprintf(format, args...)
	}
}

func (c colorType) Printf(format string, args ...interface{}) {
	c.Print(fmt.Sprintf(format, args...))
}

func (c colorType) PrintfLn(format string, args ...interface{}) {
	c.Println(fmt.Sprintf(format, args...))
}
