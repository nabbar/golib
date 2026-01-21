/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package hooksyslog

/*
RFC 5424 - The Syslog Protocol - 6.2.1 :

	The PRI part MUST have three, four, or five characters and will be
	bound with angle brackets as the first and last characters.  The PRI
	part starts with a leading "<" ('less-than' character, %d60),
	followed by a number, which is followed by a ">" ('greater-than'
	character, %d62).  The number contained within these angle brackets
	is known as the Priority value (PRIVAL) and represents both the
	Facility and Severity.  The Priority value consists of one, two, or
	three decimal integers (ABNF DIGITS) using values of %d48 (for "0")
	through %d57 (for "9").
	The Priority value is calculated by first multiplying the Facility
	number by 8 and then adding the numerical value of the Severity.  For
	example, a kernel message (Facility=0) with a Severity of Emergency
	(Severity=0) would have a Priority value of 0.  Also, a "local use 4"
	message (Facility=20) with a Severity of Notice (Severity=5) would
	have a Priority value of 165.  In the PRI of a syslog message, these
	values would be placed between the angle brackets as <0> and <165>
	respectively.  The only time a value of "0" follows the "<" is for
	the Priority value of "0".  Otherwise, leading "0"s MUST NOT be used.
*/

// PriorityCalc calculates the syslog Priority value (PRIVAL) as defined in RFC 5424.
// PRIVAL = (Facility * 8) + Severity.
// This value is placed at the beginning of a syslog message in angle brackets (e.g., <165>).
func PriorityCalc(f Facility, s Severity) uint8 {
	// move 3 bits same as multiplying by 8
	return (f.Uint8() << 3) | s.Uint8()
}
