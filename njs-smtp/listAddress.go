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

import "strings"

type listMailAddress map[int]MailAddress

func (lst listMailAddress) Len() int {
	return len(lst)
}

func (lst listMailAddress) IsEmpty() bool {
	return len(lst) < 1
}

func (lst listMailAddress) Merge(list ListMailAddress) {
	if list == nil {
		return
	}

	for _, adr := range list.Slice() {
		lst[lst.Len()] = adr
	}
}

func (lst listMailAddress) Slice() []MailAddress {
	var res = make([]MailAddress, 0)

	for _, adr := range lst {
		res = append(res, adr)
	}

	return res
}

func (lst listMailAddress) Add(adr ...MailAddress) {
	for _, a := range adr {
		lst[lst.Len()] = a
	}
}

func (lst listMailAddress) AddParseEmail(m ...string) {
	for _, a := range m {
		lst.Add(MailAddressParser(a))
	}
}

func (lst listMailAddress) AddNewEmail(name, addr string) {
	lst.Add(NewMailAddress(name, addr))
}

func (lst listMailAddress) AddressOnly() string {
	var res = make([]string, 0)

	for _, m := range lst {
		adr := m.AddressOnly()
		if adr != "" {
			res = append(res, adr)
		}
	}

	return strings.Join(res, ",")
}

func (lst listMailAddress) String() string {
	var res = make([]string, 0)

	for _, m := range lst {
		adr := m.String()
		if adr != "" {
			res = append(res, adr)
		}
	}

	return strings.Join(res, ",")
}

func (lst listMailAddress) Clone() ListMailAddress {
	var l = make(listMailAddress)

	for _, a := range lst {
		l.Add(a.Clone())
	}

	return l
}

type ListMailAddress interface {
	Len() int
	IsEmpty() bool
	String() string
	Clone() ListMailAddress

	Merge(list ListMailAddress)
	Slice() []MailAddress

	AddressOnly() string

	Add(adr ...MailAddress)
	AddParseEmail(m ...string)
	AddNewEmail(name, addr string)
}

func NewListMailAddress() ListMailAddress {
	return make(listMailAddress)
}
