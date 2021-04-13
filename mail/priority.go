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

package mail

const (
	headerImportance     = "Importance"
	headerMSMailPriority = "X-MSMail-Priority"
	headerPriority       = "X-Priority"
)

type Priority uint8

const (
	// PriorityNormal sets the email priority to normal.
	PriorityNormal Priority = iota
	// PriorityLow sets the email priority to Low.
	PriorityLow
	// PriorityHigh sets the email priority to High.
	PriorityHigh
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "Low"
	case PriorityHigh:
		return "High"
	case PriorityNormal:
		return "Normal"
	}

	return PriorityNormal.String()
}

func (p Priority) headerPriority() string {
	switch p {
	case PriorityLow:
		return "5 (Lowest)"
	case PriorityHigh:
		return "1 (Highest)"
	case PriorityNormal:
		return ""
	}

	return PriorityNormal.headerPriority()
}

func (p Priority) headerImportance() string {
	switch p {
	case PriorityLow:
		return PriorityLow.String()
	case PriorityHigh:
		return PriorityHigh.String()
	case PriorityNormal:
		return ""
	}

	return PriorityNormal.headerImportance()
}

func (p Priority) headerMSMailPriority() string {
	switch p {
	case PriorityLow:
		return PriorityLow.String()
	case PriorityHigh:
		return PriorityHigh.String()
	case PriorityNormal:
		return ""
	}

	return PriorityNormal.headerMSMailPriority()
}

func (p Priority) getHeader(h func(key string, values ...string)) {
	for k, f := range map[string]func() string{
		headerPriority:       p.headerPriority,
		headerMSMailPriority: p.headerMSMailPriority,
		headerImportance:     p.headerImportance,
	} {
		if v := f(); k != "" && v != "" {
			h(k, v)
		}
	}
}
