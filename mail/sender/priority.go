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

import "strings"

const (
	// Email priority header names used by various email clients
	headerImportance     = "Importance"        // Standard importance header (RFC 2156)
	headerMSMailPriority = "X-MSMail-Priority" // Microsoft-specific priority header
	headerPriority       = "X-Priority"        // X-Priority header (widely supported)
)

// Priority defines the urgency level of an email message.
//
// Email priority is indicated through multiple headers to ensure compatibility
// with different email clients. The priority setting affects how the email
// is displayed to recipients but does NOT affect delivery speed or routing.
//
// Priority levels help recipients identify which emails need immediate attention:
//   - High: Urgent messages requiring immediate action
//   - Normal: Standard messages (default)
//   - Low: Non-urgent messages that can be handled later
//
// Example usage:
//
//	mail.SetPriority(sender.PriorityHigh)   // For urgent notifications
//	mail.SetPriority(sender.PriorityNormal) // For standard communications
//	mail.SetPriority(sender.PriorityLow)    // For newsletters or bulk emails
//
// Note: Priority is a suggestion to email clients; the actual treatment
// depends on the recipient's email client and settings.
type Priority uint8

const (
	// PriorityNormal represents standard email priority (default).
	// This is the most commonly used priority level and should be used for
	// regular business correspondence and standard notifications.
	//
	// When set to Normal, priority headers are typically omitted from the email,
	// allowing the email client to use its default display behavior.
	//
	// Example:
	//	mail.SetPriority(sender.PriorityNormal)
	PriorityNormal Priority = iota

	// PriorityLow represents low priority email.
	// Use this for non-urgent messages such as:
	//   - Newsletters and marketing materials
	//   - Automated reports that don't require immediate attention
	//   - Bulk notifications
	//   - Informational updates
	//
	// Low priority emails are often displayed with a down arrow or other indicator
	// in email clients, signaling to recipients that they can be handled later.
	//
	// Email headers set:
	//   - X-Priority: 5 (Lowest)
	//   - Importance: Low
	//   - X-MSMail-Priority: Low
	//
	// Example:
	//	mail.SetPriority(sender.PriorityLow)
	PriorityLow

	// PriorityHigh represents high priority email.
	// Use this for urgent messages such as:
	//   - Critical system alerts
	//   - Time-sensitive notifications
	//   - Emergency communications
	//   - Deadline reminders
	//
	// High priority emails are typically displayed with an exclamation mark or
	// other visual indicator in email clients.
	//
	// Important: Use high priority sparingly. Overuse can lead to recipients
	// ignoring priority flags or filtering your emails.
	//
	// Email headers set:
	//   - X-Priority: 1 (Highest)
	//   - Importance: High
	//   - X-MSMail-Priority: High
	//
	// Example:
	//	mail.SetPriority(sender.PriorityHigh)
	PriorityHigh
)

// String returns a human-readable string representation of the Priority.
//
// Returns:
//   - "Normal" for PriorityNormal
//   - "Low" for PriorityLow
//   - "High" for PriorityHigh
//   - Defaults to "Normal" for unknown values
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

// headerPriority returns the value for the X-Priority header.
// This header uses numeric values: 1 (Highest) to 5 (Lowest).
//
// Returns:
//   - "5 (Lowest)" for PriorityLow
//   - "1 (Highest)" for PriorityHigh
//   - "" (empty) for PriorityNormal (header is omitted)
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

// headerImportance returns the value for the Importance header (RFC 2156).
// This is a standard header recognized by most email clients.
//
// Returns:
//   - "Low" for PriorityLow
//   - "High" for PriorityHigh
//   - "" (empty) for PriorityNormal (header is omitted)
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

// headerMSMailPriority returns the value for the X-MSMail-Priority header.
// This is a Microsoft-specific header used by Outlook and other MS email clients.
//
// Returns:
//   - "Low" for PriorityLow
//   - "High" for PriorityHigh
//   - "" (empty) for PriorityNormal (header is omitted)
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

// getHeader calls the provided function with all priority-related headers.
// This is used internally to add priority headers to the email.
//
// The function adds multiple headers for maximum compatibility across
// different email clients (Outlook, Thunderbird, Gmail, etc.).
//
// Parameters:
//   - h: A callback function that receives header key-value pairs
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

// ParsePriority converts a string representation into a Priority value.
// The comparison is case-insensitive for flexibility.
//
// Parameters:
//   - s: String representation of the priority. Valid values are:
//     "Normal", "Low", "High" (case-insensitive)
//
// Returns:
//   - The corresponding Priority value
//   - PriorityNormal if the string doesn't match any known priority
//
// Example:
//
//	priority := sender.ParsePriority("High")      // Returns PriorityHigh
//	priority := sender.ParsePriority("high")      // Also returns PriorityHigh
//	priority := sender.ParsePriority("unknown")   // Returns PriorityNormal
func ParsePriority(s string) Priority {
	switch strings.ToUpper(s) {
	case strings.ToUpper(PriorityLow.String()):
		return PriorityLow
	case strings.ToUpper(PriorityHigh.String()):
		return PriorityHigh
	default:
		return PriorityNormal
	}
}
