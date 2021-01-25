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

package ldap

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/nabbar/golib/errors"
)

type TLSMode uint8

const (
	tlsmode_init TLSMode = iota
	//TLSModeNone no tls connection.
	TLSMODE_NONE TLSMode = iota + 1
	//TLSModeTLS strict tls connection.
	TLSMODE_TLS
	//TLSModeStartTLS starttls connection (tls into a no tls connection).
	TLSMODE_STARTTLS
)

const (
	groupFieldCN = "cn"
	userFieldUid = "uid"
)

func (m TLSMode) String() string {
	switch m {
	case TLSMODE_STARTTLS:
		return "starttls"
	case TLSMODE_TLS:
		return "tls"
	case TLSMODE_NONE:
		return "none"
	case tlsmode_init:
		return "not defined"
	}

	return tlsmode_init.String()
}

func GetDefaultAttributes() []string {
	return []string{"givenName", "mail", "uid", "dn"}
}

type Config struct {
	Uri       string `cloud:"uri" mapstructure:"uri" json:"uri" yaml:"uri" toml:"uri" validate:"fqdn,required"`
	PortLdap  int    `cloud:"port-ldap" mapstructure:"port-ldap" json:"port-ldap" yaml:"port-ldap" toml:"port-ldap" validate:"number,gte=0,nefield=Portldaps,required"`
	Portldaps int    `cloud:"port-ldaps" mapstructure:"port-ldaps" json:"port-ldaps" yaml:"port-ldaps" toml:"port-ldaps" validate:"number,nefield=Portldap,omitempty"`
	Basedn    string `cloud:"basedn" mapstructure:"basedn" json:"basedn" yaml:"basedn" toml:"basedn" validate:"printascii,omitempty"`
	//FilterGroup is fmt pattern like '(&(objectClass=groupOfNames)(%s=%s))' to make search of group object class
	FilterGroup string `cloud:"filter-group" mapstructure:"filter-group" json:"filter-group" yaml:"filter-group" toml:"filter-group" validate:"printascii,required"`
	//FilterUser is a fmt pattern like '(%s=%s)' to make search of user. By default, uid field is 'uid'
	FilterUser string `cloud:"filter-user" mapstructure:"filter-user" json:"filter-user" yaml:"filter-user" toml:"filter-user" validate:"printascii,required"`
}

func NewConfig() *Config {
	return &Config{}
}

func (cnf Config) Clone() *Config {
	return &Config{
		Uri:         cnf.Uri,
		PortLdap:    cnf.PortLdap,
		Portldaps:   cnf.Portldaps,
		Basedn:      cnf.Basedn,
		FilterGroup: cnf.FilterGroup,
		FilterUser:  cnf.FilterUser,
	}
}

func (cnf Config) BaseDN() string {
	return cnf.Basedn
}

func (cnf Config) ServerAddr(withTls bool) string {
	if withTls {
		return fmt.Sprintf("%s:%d", cnf.Uri, cnf.Portldaps)
	}

	return fmt.Sprintf("%s:%d", cnf.Uri, cnf.PortLdap)
}

func (cnf Config) PatternFilterGroup() string {
	return cnf.FilterGroup
}

func (cnf Config) PatternFilterUser() string {
	return cnf.FilterUser
}

func (cnf Config) Validate() errors.Error {
	var e = ErrorLDAPValidatorError.Error(nil)

	if err := validator.New().Struct(cnf); err != nil {
		if er, ok := err.(*validator.InvalidValidationError); ok {
			e.AddParent(er)
		}

		for _, err := range err.(validator.ValidationErrors) {
			e.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", err.StructNamespace(), err.ActualTag()))
		}
	}

	if !e.HasParent() {
		e = nil
	}

	return e
}
