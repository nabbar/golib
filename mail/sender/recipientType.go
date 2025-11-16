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

package sender

// recipientType defines the different categories of email recipients.
//
// Email recipients can be classified into three main categories, each with
// different visibility and intent:
//   - To: Primary recipients who are expected to act on the email
//   - CC: Secondary recipients who should be kept informed
//   - BCC: Hidden recipients who receive the email without other recipients knowing
//
// Example usage:
//
//	mail.Email().AddRecipients(sender.RecipientTo, "user@example.com")
//	mail.Email().AddRecipients(sender.RecipientCC, "manager@example.com")
//	mail.Email().AddRecipients(sender.RecipientBCC, "archive@example.com")
type recipientType uint8

const (
	// RecipientTo represents primary recipients of the email.
	// These recipients appear in the "To" field of the email header and are
	// typically the main audience who should act on or respond to the email.
	//
	// All To recipients can see each other's email addresses in the email header.
	//
	// Example:
	//	mail.Email().AddRecipients(sender.RecipientTo, "user@example.com", "team@example.com")
	RecipientTo recipientType = iota

	// RecipientCC represents "Carbon Copy" recipients.
	// These recipients appear in the "Cc" field of the email header and are
	// secondary recipients who should be kept informed but are not the primary
	// audience. They typically don't need to take action on the email.
	//
	// All recipients (To and CC) can see CC addresses in the email header.
	//
	// Example:
	//	mail.Email().AddRecipients(sender.RecipientCC, "manager@example.com")
	RecipientCC

	// RecipientBCC represents "Blind Carbon Copy" recipients.
	// These recipients receive the email but their addresses are NOT visible
	// to any other recipients. This is useful for:
	//   - Protecting recipient privacy in bulk emails
	//   - Sending copies to archives without other recipients knowing
	//   - Including supervisors without making their oversight obvious
	//
	// BCC recipients can only see themselves and the To/CC recipients.
	// Other recipients cannot see that BCC recipients exist.
	//
	// Example:
	//	mail.Email().AddRecipients(sender.RecipientBCC, "archive@example.com", "audit@example.com")
	RecipientBCC
)

// String returns a human-readable string representation of the recipientType.
//
// Returns:
//   - "To" for RecipientTo
//   - "Cc" for RecipientCC  (note: proper capitalization for email headers)
//   - "Bcc" for RecipientBCC (note: proper capitalization for email headers)
//   - Defaults to "To" for unknown values
//
// The returned strings match standard email header field names.
func (r recipientType) String() string {
	switch r {
	case RecipientTo:
		return "To"
	case RecipientCC:
		return "Cc"
	case RecipientBCC:
		return "Bcc"
	}

	return RecipientTo.String()
}
