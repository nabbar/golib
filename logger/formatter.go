/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package logger

import (
	"io"
	"strings"

	"github.com/sirupsen/logrus"
)

// Format a uint8 type customized with function to manage the result logger format
type Format uint8

const (
	nilFormat Format = iota
	// TextFormat a text format for logger entry
	TextFormat
	// JsonFormat a json format for logger entry
	JsonFormat
)

var (
	curFormat = TextFormat
)

func SetOutput(out io.WriteCloser) {
	logrus.SetOutput(out)
}

func updateFormatter(newFormat Format) {
	if newFormat != nilFormat {
		curFormat = newFormat
	}

	switch curFormat {
	case TextFormat:
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:            modeColor,
			DisableColors:          !modeColor,
			DisableLevelTruncation: !modeColor,
			DisableTimestamp:       true,
			DisableSorting:         true,
		})
	case JsonFormat:
		logrus.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp: true,
		})
	}
}

// GetFormatListString return the full list (slice of string) of all available formats
func GetFormatListString() []string {
	return []string{
		strings.ToLower(TextFormat.String()),
		strings.ToLower(JsonFormat.String()),
	}
}

// SetFormat Change the format of all log entry with the Format type given in parameter. The change is apply for next entry only
//
// If the given Format type is not matching a correct Format type, no change will be apply.
/*
	fmt a Format type for the format to use
*/
func SetFormat(fmt Format) {
	switch fmt {
	case TextFormat, JsonFormat:
		updateFormatter(fmt)
	}
}

// GetCurrentFormat Return the current Format Type used for all log entry
func GetCurrentFormat() Format {
	return curFormat
}

// GetFormatString return a valid Format Type matching the given string parameter
/*
	format the string representation of a Format type
*/
func GetFormatString(format string) Format {
	switch strings.ToLower(format) {
	case strings.ToLower(TextFormat.String()):
		return TextFormat

	case strings.ToLower(JsonFormat.String()):
		return JsonFormat

	default:
		return TextFormat
	}
}

// String Return the string name of the Format Type
func (f Format) String() string {
	switch f {
	case JsonFormat:
		return "Json"
	default:
		return "Text"
	}
}
