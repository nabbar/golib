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

package njs_ldap_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/nabbar/golib/njs-ldap"
	"github.com/nabbar/golib/njs-ldap/ldaptestserver"
	"github.com/nabbar/golib/njs-logger"
)

var (
	username   = ""
	grpAdmin   = "Test - Group2"
	grpManager = "Test - Group3"
	userBindEr = "uid=false,dc=example,dc=com"
	userBindDn = "uid=" + ldaptestserver.GetTestAccountDN(ldaptestserver.TEST_ACCOUNT_TYPE_USER, ldaptestserver.TEST_ACCOUNT_ID_MAIN) + ",dc=example,dc=com"
	userBindPw = "abc123def"
)

func init() {
	njs_logger.SetLevel(njs_logger.InfoLevel)
}

var _ = Describe("Helpers Test: ", func() {
	Context("Test LDAP", func() {
		Context("having a false user connection", func() {
			BeforeEach(func() {
				ldap = njs_ldap.NewLDAP(conf, njs_ldap.GetDefaultAttributes())
				ldap.SetCredentials(userBindEr, userBindPw)
			})
			It("have a connected instance of helper ldap", func() {
				Expect(ldap).ToNot(BeNil())
				Expect(ldap.Connect()).To(HaveOccurred())
			})
		})
		Context("having a good user connection", func() {
			BeforeEach(func() {
				ldap = njs_ldap.NewLDAP(conf, njs_ldap.GetDefaultAttributes())
				ldap.SetCredentials(userBindDn, userBindPw)
			})
			It("have a connected instance of helper ldap", func() {
				Expect(ldap).ToNot(BeNil())
				Expect(ldap.Connect()).ToNot(HaveOccurred())
			})

			It("return an error on retrieve User Info of a false username", func() {
				usr, err := ldap.UserInfo("notExists")
				Expect(err).To(HaveOccurred())
				Expect(usr).To(Equal(make(map[string]string)))
			})
		})
		Context("given a real username", func() {
			BeforeEach(func() {
				username = "Test2"
				ldap = njs_ldap.NewLDAP(conf, njs_ldap.GetDefaultAttributes())
				ldap.SetCredentials(userBindDn, userBindPw)
			})

			It("have a connected instance of helper ldap", func() {
				Expect(ldap).ToNot(BeNil())
				Expect(ldap.Connect()).ToNot(HaveOccurred())
			})

			It("return the retrieved User Info", func() {
				usr, err := ldap.UserInfo("Test2")
				Expect(err).ToNot(HaveOccurred())
				Expect(usr).ToNot(Equal(make(map[string]string)))
				Expect(usr["uid"]).To(Equal("Test2"))
			})

			It("Return true and no error when check Is in Admin group", func() {
				chk, err := ldap.UserIsInGroup(username, []string{grpAdmin})
				Expect(err).ToNot(HaveOccurred())
				Expect(chk).To(BeTrue())
			})

			It("Return false and no error when check Is in Manager group", func() {
				chk, err := ldap.UserIsInGroup(username, []string{grpManager})
				Expect(err).ToNot(HaveOccurred())
				Expect(chk).To(BeFalse())
			})

			It("Return no error and the list when retrieve members of a group", func() {
				lst, err := ldap.UsersOfGroup("Test - Group1")
				Expect(err).ToNot(HaveOccurred())
				Expect(len(lst)).To(Equal(3))
			})

			Context("given a empty password", func() {
				It("return an error on authentificate with LDAP", func() {
					Expect(ldap.AuthUser("uid="+username, "")).To(HaveOccurred())
				})
			})

			Context("given a wrong password", func() {
				It("return no error but a false result when authentificate with LDAP", func() {
					Expect(ldap.AuthUser("uid="+username, "wrongPassword")).To(HaveOccurred())
				})
			})

			Context("given a good password", func() {
				It("return no error but a false result when authentificate with LDAP", func() {
					Expect(ldap.AuthUser("uid="+username, "abc123def")).ToNot(HaveOccurred())
				})
			})
		})

	})
})
