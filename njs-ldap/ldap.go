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

package njs_ldap

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/nabbar/golib/njs-certif"
	"github.com/nabbar/golib/njs-logger"
	"gopkg.in/ldap.v3"
)

//HelperLDAP struct use to manage connection to server and request it
type HelperLDAP struct {
	Attributes []string
	conn       *ldap.Conn
	config     *Config
	tlsConfig  *tls.Config
	tlsMode    TLSMode
	bindDN     string
	bindPass   string
}

//NewLDAP build a new LDAP helper based on config struct given
func NewLDAP(cnf *Config, attributes []string) *HelperLDAP {
	if cnf == nil {
		panic("given config is a nil struct")
	}

	return &HelperLDAP{
		Attributes: attributes,
		tlsConfig:  njs_certif.GetTLSConfig(cnf.Uri),
		tlsMode:    tlsmode_init,
		config:     cnf.Clone(),
	}
}

//SetCredentials used to defined the BindDN and password for connection
func (lc *HelperLDAP) SetCredentials(user, pass string) {
	lc.bindDN = user
	lc.bindPass = pass
}

//SetCredentials used to defined the BindDN and password for connection
func (lc *HelperLDAP) ForceTLSMode(tlsMode TLSMode, tlsConfig *tls.Config) {
	switch tlsMode {
	case TLSMODE_TLS, TLSMODE_STARTTLS, TLSMODE_NONE:
		lc.tlsConfig = tlsConfig
	}

	if tlsConfig != nil {
		lc.tlsConfig = tlsConfig
	}
}

func (lc *HelperLDAP) tryConnect() (TLSMode, error) {
	var (
		l   *ldap.Conn
		err error
	)

	defer func(l *ldap.Conn) {
		if l != nil {
			l.Close()
		}
	}(l)

	if lc.config.Portldaps != 0 {
		l, err = ldap.DialTLS("tcp", lc.config.ServerAddr(true), lc.tlsConfig)
		if err == nil {
			njs_logger.DebugLevel.Logf("ldap connected with tls mode '%s'", lc.tlsMode.String())
			return TLSMODE_TLS, nil
		}
	}

	if lc.config.PortLdap == 0 {
		return 0, fmt.Errorf("ldap server not well defined")
	}

	l, err = ldap.Dial("tcp", lc.config.ServerAddr(false))
	if err != nil {
		return 0, err
	}

	if e := l.StartTLS(lc.tlsConfig); e == nil {
		njs_logger.DebugLevel.Logf("ldap connected with tls mode '%s'", lc.tlsMode.String())
		return TLSMODE_STARTTLS, nil
	}

	njs_logger.DebugLevel.Logf("ldap connected with tls mode '%s'", lc.tlsMode.String())
	return TLSMODE_NONE, nil
}

func (lc *HelperLDAP) connect() error {
	if lc.conn == nil {
		var (
			l   *ldap.Conn
			err error
		)

		if lc.tlsMode == tlsmode_init {
			m, e := lc.tryConnect()

			if e != nil {
				return e
			}

			lc.tlsMode = m
		}

		if lc.tlsMode == TLSMODE_TLS {
			l, err = ldap.DialTLS("tcp", lc.config.ServerAddr(true), lc.tlsConfig)
			if err != nil {
				return fmt.Errorf("ldap connection error with tls mode '%s': %v", lc.tlsMode.String(), err)
			}
		}

		if lc.tlsMode == TLSMODE_NONE || lc.tlsMode == TLSMODE_STARTTLS {
			l, err = ldap.Dial("tcp", lc.config.ServerAddr(false))
			if err != nil {
				return fmt.Errorf("ldap connection error with tls mode '%s': %v", lc.tlsMode.String(), err)
			}
		}

		if lc.tlsMode == TLSMODE_STARTTLS {
			err = l.StartTLS(lc.tlsConfig)
			if err != nil {
				return fmt.Errorf("ldap connection error with tls mode '%s': %v", lc.tlsMode.String(), err)
			}
		}

		njs_logger.DebugLevel.Logf("ldap connected with tls mode '%s'", lc.tlsMode.String())
		lc.conn = l
	}

	return nil
}

//Check used to check if connection success (without any bind)
func (lc *HelperLDAP) Check() error {
	if err := lc.connect(); err != nil {
		return err
	}

	lc.Close()
	return nil
}

//Close used to close connection object
func (lc *HelperLDAP) Close() {
	if lc.conn != nil {
		lc.conn.Close()
		lc.conn = nil
	}
}

//AuthUser used to test bind given user uid and password
func (lc *HelperLDAP) AuthUser(username, password string) error {

	if err := lc.connect(); err != nil {
		return err
	}

	if username == "" || password == "" {
		return errors.New("Cannot bind with partial credentials, bindDN or bind password is empty string")
	}

	return lc.conn.Bind(username, password)
}

//Connect used to connect and bind to server
func (lc *HelperLDAP) Connect() error {
	if err := lc.AuthUser(lc.bindDN, lc.bindPass); err != nil {
		return fmt.Errorf("error while trying to bind on LDAP server %s with tls mode '%s': %v", lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), err)
	}

	njs_logger.DebugLevel.Logf("Bind success on LDAP server %s with tls mode '%s'", lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String())
	return nil
}

func (lc *HelperLDAP) runSearch(filter string, attributes []string) (*ldap.SearchResult, error) {
	var (
		err error
		src *ldap.SearchResult
	)

	if err = lc.Connect(); err != nil {
		return nil, err
	}

	defer lc.Close()

	searchRequest := ldap.NewSearchRequest(
		lc.config.Basedn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		100, 0, false,
		filter,
		attributes,
		nil,
	)

	if src, err = lc.conn.Search(searchRequest); err != nil {
		return nil, fmt.Errorf("error while looking for '%s' on ldap server %s with tls mode '%s': %v", filter, lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), err)
	}

	njs_logger.DebugLevel.Logf("Search success on server '%s' with tls mode '%s', with filter [%s] and attribute %v", lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), filter, attributes)
	return src, nil
}

//UserInfo used to retrieve the information of a given username
func (lc *HelperLDAP) UserInfo(username string) (map[string]string, error) {
	var (
		err     error
		src     *ldap.SearchResult
		userRes map[string]string
	)

	if username == "" {
		usr := lc.ParseEntries(lc.bindDN)
		username = usr["uid"][0]
	}

	userRes = make(map[string]string)
	attributes := append(lc.Attributes, "cn")

	if src, err = lc.runSearch(fmt.Sprintf(lc.config.FilterUser, username), attributes); err != nil {
		return userRes, err
	}

	if len(src.Entries) != 1 {
		if len(src.Entries) > 1 {
			err = errors.New("Username not unique")
		} else {
			err = errors.New("Username not found")
		}

		return userRes, fmt.Errorf("error while looking for username '%s' on ldap server '%s' with tls mode '%s': %v", username, lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), err)
	}

	for _, attr := range attributes {
		userRes[attr] = src.Entries[0].GetAttributeValue(attr)
	}

	if _, ok := userRes["DN"]; !ok {
		userRes["DN"] = src.Entries[0].DN
	}

	njs_logger.DebugLevel.Logf("Map info retrieve in ldap server '%s' with tls mode '%s' about user [%s] : %v", lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), username, userRes)
	return userRes, nil
}

//UserMemberOf returns the group list of a given user.
func (lc *HelperLDAP) UserMemberOf(username string) ([]string, error) {
	var (
		err error
		src *ldap.SearchResult
		grp []string
	)

	if username == "" {
		usr := lc.ParseEntries(lc.bindDN)
		username = usr["uid"][0]
	}

	grp = make([]string, 0)

	if src, err = lc.runSearch(fmt.Sprintf(lc.config.FilterUser, username), []string{"memberOf"}); err != nil {
		return grp, err
	}

	for _, entry := range src.Entries {
		for _, mmb := range entry.GetAttributeValues("memberOf") {
			njs_logger.DebugLevel.Logf("Group find for uid '%s' on server '%s' with tls mode '%s' : %v", username, lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), mmb)
			mmo := lc.ParseEntries(mmb)
			grp = append(grp, mmo["cn"]...)
		}
	}

	njs_logger.DebugLevel.Logf("Groups find for uid '%s' on server '%s' with tls mode '%s' : %v", username, lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), grp)
	return grp, nil
}

//UserIsInGroup used to check if a given username is a group member of a list of reference group name
func (lc *HelperLDAP) UserIsInGroup(username string, groupname []string) (bool, error) {
	var (
		err     error
		grpMmbr []string
	)

	if username == "" {
		usr := lc.ParseEntries(lc.bindDN)
		username = usr["uid"][0]
	}

	if grpMmbr, err = lc.UserMemberOf(username); err != nil {
		return false, err
	}

	for _, grpSrch := range groupname {
		for _, grpItem := range grpMmbr {
			if strings.ToUpper(grpSrch) == strings.ToUpper(grpItem) {
				return true, nil
			}
		}
	}

	return false, nil
}

//UsersOfGroup used to retrieve the member list of a given group name
func (lc *HelperLDAP) UsersOfGroup(groupname string) ([]string, error) {
	var (
		err error
		src *ldap.SearchResult
		grp []string
	)

	grp = make([]string, 0)

	if src, err = lc.runSearch(fmt.Sprintf(lc.config.FilterGroup, groupname), []string{"member"}); err != nil {
		return grp, err
	}

	for _, entry := range src.Entries {
		for _, mmb := range entry.GetAttributeValues("member") {
			member := lc.ParseEntries(mmb)
			grp = append(grp, member["uid"]...)
		}
	}

	njs_logger.DebugLevel.Logf("Member of groups [%s] find on server '%s' with tls mode '%s' : %v", groupname, lc.config.ServerAddr(lc.tlsMode == TLSMODE_TLS), lc.tlsMode.String(), grp)
	return grp, nil
}

//ParseEntries used to clean attributes of an object class
func (lc HelperLDAP) ParseEntries(entry string) map[string][]string {
	var listEntries = make(map[string][]string)

	for _, ent := range strings.Split(entry, ",") {
		key := strings.SplitN(ent, "=", 2)

		if len(key) != 2 || len(key[0]) < 1 || len(key[1]) < 1 {
			continue
		}

		key[0] = strings.TrimSpace(key[0])
		key[1] = strings.TrimSpace(key[1])

		if _, ok := listEntries[key[0]]; !ok {
			listEntries[key[0]] = []string{}
		}

		listEntries[key[0]] = append(listEntries[key[0]], key[1])
	}

	return listEntries
}
