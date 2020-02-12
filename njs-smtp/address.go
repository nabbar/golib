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

package njs_smtp

import (
	"net/mail"
	"strings"
)

type mailAddress struct {
	mail.Address
}

func (adr *mailAddress) ForceName(name string) {
	adr.Address.Name = name
}

func (adr *mailAddress) ForceAddress(address string) {
	adr.Address.Address = address
}

func (adr mailAddress) String() string {
	return strings.TrimSpace(adr.Address.String())
}

func (adr mailAddress) AddressOnly() string {
	str := strings.TrimSpace(adr.Address.String())

	if adr.Address.Name == "" || adr.Address.Address == "" || strings.HasPrefix(str, "<") {
		str = strings.Replace(str, "\n", "", -1)
		str = strings.Trim(str, ">")
		str = strings.Trim(str, "@")
		str = strings.Trim(str, "<")
		str = strings.TrimSpace(str)
		str = strings.Trim(str, "\"")
	} else {
		str = strings.TrimSpace(adr.Address.Address)
		str = strings.Replace(str, "\n", "", -1)
	}

	return str
}

func (adr mailAddress) Clone() MailAddress {
	return &mailAddress{
		mail.Address{
			Name:    adr.Address.Name,
			Address: adr.Address.Address,
		},
	}
}

type MailAddress interface {
	Clone() MailAddress

	ForceName(name string)
	ForceAddress(address string)

	String() string
	AddressOnly() string
}

func MailAddressParser(str string) MailAddress {
	obj, err := mail.ParseAddress(str)

	if err != nil {
		obj = &mail.Address{
			Name: str,
		}
	}

	return &mailAddress{
		*obj,
	}
}

func NewMailAddress(name, address string) MailAddress {
	return &mailAddress{
		mail.Address{
			Name:    name,
			Address: address,
		},
	}
}
