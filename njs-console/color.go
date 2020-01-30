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
	"fmt"

	"github.com/fatih/color"
)

type colorType uint8

const (
	ColorPrompt colorType = iota
	ColorPrint
	ColorSPrintF
	ColorFatal
	ColorError
	ColorWarn
	ColorInfo
	ColorDebug
)

var (
	colorList map[colorType]*color.Color
)

func init() {
	colorList = map[colorType]*color.Color{
		ColorPrompt:  nil,
		ColorPrint:   nil,
		ColorSPrintF: nil,
		ColorFatal:   nil,
		ColorError:   nil,
		ColorWarn:    nil,
		ColorInfo:    nil,
		ColorDebug:   nil,
	}
}

func (c colorType) SetColor(col *color.Color) {
	colorList[c] = col
}

func (c colorType) println(text string) {
	if colorList[c] != nil {
		_, _ = colorList[c].Println(text) // #nosec
	} else {
		println(text)
	}
}

func (c colorType) print(text string) {
	if colorList[c] != nil {
		_, _ = colorList[c].Print(text) // #nosec
	} else {
		print(text)
	}
}

func (c colorType) sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func (c colorType) printf(format string, args ...interface{}) {
	c.println(fmt.Sprintf(format, args...))
}

func (c colorType) printfLn(format string, args ...interface{}) {
	c.println(fmt.Sprintf(format, args...))
}

func SPrintf(format string, args ...interface{}) string {
	return ColorSPrintF.sprintf(format, args...)
}

func SPrintfCol(col *color.Color, format string, args ...interface{}) string {
	c := colorList[ColorSPrintF]

	defer func() {
		colorList[ColorSPrintF] = c
	}()

	return SPrintf(format, args...)
}

func Print(format string, args ...interface{}) {
	ColorPrint.printf(format, args...)
}

func PrintLn(format string, args ...interface{}) {
	ColorPrint.printfLn(format, args...)
}

func Debug(text string) {
	ColorDebug.print(text)
}

func Info(text string) {
	ColorInfo.print(text)
}

func Warn(text string) {
	ColorWarn.print(text)
}

func Error(text string) {
	ColorError.print(text)
}

func Fatal(text string) {
	ColorFatal.print(text)
}
