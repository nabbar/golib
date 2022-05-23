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
	ErrorEmptyParams errors.CodeError = iota + errors.MinPkgLDAP
	ErrorLDAPContext
	ErrorLDAPServerConfig
	ErrorLDAPServerConnection
	ErrorLDAPServerDial
	ErrorLDAPServerDialClosing
	ErrorLDAPServerTLS
	ErrorLDAPServerStartTLS
	ErrorLDAPBind
	ErrorLDAPSearch
	ErrorLDAPUserNotUniq
	ErrorLDAPUserNotFound
	ErrorLDAPInvalidDN
	ErrorLDAPInvalidUID
	ErrorLDAPAttributeNotFound
	ErrorLDAPAttributeEmpty
	ErrorLDAPValidatorError
	ErrorLDAPGroupNotFound
)

var isCodeError = errors.ExistInMapMessage(ErrorEmptyParams)

func IsCodeError() bool {
	return isCodeError
}

func init() {
	errors.RegisterIdFctMessage(ErrorEmptyParams, getMessage)
}

func getMessage(code errors.CodeError) (message string) {
	switch code {
	case errors.UNK_ERROR:
		return ""
	case ErrorEmptyParams:
		return "given parameters is empty"
	case ErrorLDAPContext:
		return "LDAP server connection context occurs an error"
	case ErrorLDAPServerConfig:
		return "LDAP server config is not well defined"
	case ErrorLDAPServerConnection:
		return "connection server occurs error "
	case ErrorLDAPServerDial:
		return "dialing server occurs error "
	case ErrorLDAPServerDialClosing:
		return "dialing server is going to be closed"
	case ErrorLDAPServerTLS:
		return "cannot start dial to server with TLS Mode"
	case ErrorLDAPServerStartTLS:
		return "cannot init starttls mode on opening server connection"
	case ErrorLDAPBind:
		return "error on binding user/pass"
	case ErrorLDAPSearch:
		return "error on calling search on connected server"
	case ErrorLDAPUserNotUniq:
		return "user uid is not uniq"
	case ErrorLDAPUserNotFound:
		return "user uid is not found"
	case ErrorLDAPInvalidDN:
		return "dn given is not valid"
	case ErrorLDAPInvalidUID:
		return "uid is not found or empty"
	case ErrorLDAPAttributeNotFound:
		return "requested attribute is not found"
	case ErrorLDAPAttributeEmpty:
		return "requested attribute is empty"
	case ErrorLDAPValidatorError:
		return "invalid validation config"
	case ErrorLDAPGroupNotFound:
		return "group not found"
	}

	return ""
}
