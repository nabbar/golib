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
	"io/ioutil"
	"os"
	"testing"

	"github.com/nabbar/golib/njs-logger"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/njs-ldap"
	"github.com/nabbar/golib/njs-ldap/ldaptestserver"
)

var (
	ldap *njs_ldap.HelperLDAP
	conf *njs_ldap.Config
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	filepath := dir + "/test.yml"

	if _, err := os.Stat(filepath); err != nil {
		panic(err)
	}

	conf = njs_ldap.NewConfig()

	if cnt, err := ioutil.ReadFile(filepath); err != nil {
		panic(err)
	} else if err := yaml.Unmarshal(cnt, &conf); err != nil {
		panic(err)
	}
}

func TestHelpers(t *testing.T) {
	njs_logger.InfoLevel.Log("Starting LDAP Test Server...")
	ldaptestserver.RunTestLDAPServer()

	defer func() {
		if ldap != nil {
			ldap.Close()
		}
		ldaptestserver.StopTestLDAPServer()
		njs_logger.InfoLevel.Log("LDAP Test Server is stopped...")
	}()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite of Helpers LDAP")
}
