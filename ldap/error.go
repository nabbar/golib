/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package ldap

import "github.com/nabbar/golib/errors"

const (
	EMPTY_PARAMS errors.CodeError = iota + errors.MIN_PKG_LDAP
	LDAP_SERVER_CONFIG
	LDAP_SERVER_DIAL
	LDAP_SERVER_TLS
	LDAP_SERVER_STARTTLS
	LDAP_BIND
	LDAP_SEARCH
	LDAP_USER_NOT_UNIQ
	LDAP_USER_NOT_FOUND
)

func init() {
	errors.RegisterFctMessage(getMessage)
}

func getMessage(code errors.CodeError) (message string) {
	switch code {
	case EMPTY_PARAMS:
		return "given parameters is empty"
	case LDAP_SERVER_CONFIG:
		return "LDAP server config is not well defined"
	case LDAP_SERVER_DIAL:
		return "dialing server occurs error "
	case LDAP_SERVER_TLS:
		return "cannot start dial to server with TLS Mode"
	case LDAP_SERVER_STARTTLS:
		return "cannot init starttls mode on opening server connection"
	case LDAP_BIND:
		return "error on binding user/pass"
	case LDAP_SEARCH:
		return "error on calling search on connected server"
	case LDAP_USER_NOT_UNIQ:
		return "user uid is not uniq"
	case LDAP_USER_NOT_FOUND:
		return "user uid is not found"
	}

	return ""
}
