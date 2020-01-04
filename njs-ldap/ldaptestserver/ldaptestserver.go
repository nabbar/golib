package ldaptestserver

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/sirupsen/logrus"
	"github.com/vjeantet/ldapserver"
)

const (
	TEST_ACCOUNT_TYPE_USER = iota
	TEST_ACCOUNT_TYPE_GROUP
)

const (
	TEST_ACCOUNT_ID_MAIN = iota
	TEST_ACCOUNT_ID_1
	TEST_ACCOUNT_ID_2
	TEST_ACCOUNT_ID_3
)

const (
	TEST_ACCOUNT_VALID_PASSWORD = "abc123def"
)

var (
	ch             chan os.Signal
	testLDAPserver *ldapserver.Server
)

func GetTestAccountDN(typeAccount int, idAccound int) string {
	switch typeAccount {
	case TEST_ACCOUNT_TYPE_GROUP:
		switch idAccound {
		case TEST_ACCOUNT_ID_1:
			return "Test - Group1"
		case TEST_ACCOUNT_ID_2:
			return "Test - Group2"
		case TEST_ACCOUNT_ID_3:
			return "Test - Group3"
		}
	case TEST_ACCOUNT_TYPE_USER:
		switch idAccound {
		case TEST_ACCOUNT_ID_1:
			return "Test1"
		case TEST_ACCOUNT_ID_2:
			return "Test2"
		case TEST_ACCOUNT_ID_3:
			return "Test3"
		case TEST_ACCOUNT_ID_MAIN:
			return "bindMainReadOnly"
		}
	}

	return ""
}

func RunTestLDAPServer() {
	//Create a new LDAP Server
	testLDAPserver = ldapserver.NewServer()

	//Create routes bindings
	routes := ldapserver.NewRouteMux()
	routes.NotFound(handleNotFound)
	routes.Abandon(handleAbandon)
	routes.Bind(handleBind)
	routes.Compare(handleCompare)
	routes.Add(handleAdd)
	routes.Delete(handleDelete)
	routes.Modify(handleModify)
	routes.Extended(handleStartTLS).RequestName(ldapserver.NoticeOfStartTLS)
	routes.Extended(handleWhoAmI).RequestName(ldapserver.NoticeOfWhoAmI)

	routes.Extended(handleExtended)

	routes.Search(handleSearch)

	//Attach routes to server
	testLDAPserver.Handle(routes)

	ch = make(chan os.Signal)
	// listen on 10389 and serve
	go func() {
		defer ginkgo.GinkgoRecover()

		if err := testLDAPserver.ListenAndServe("127.0.0.1:10389"); err != nil {
			logrus.Fatal("Error on LDAP Test Server : %v", err)
		}
	}()
	time.Sleep(5 * time.Second)
}

func StopTestLDAPServer() {
	// When CTRL+C, SIGINT and SIGTERM signal occurs
	// Then stop server gracefully
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	//	<-ch
	time.Sleep(5 * time.Second)
	close(ch)
	testLDAPserver.Stop()
}

func handleNotFound(w ldapserver.ResponseWriter, r *ldapserver.Message) {
	switch r.GetProtocolOp() {
	case ldapserver.ApplicationBindRequest:
		res := ldapserver.NewBindResponse(ldapserver.LDAPResultSuccess)
		res.DiagnosticMessage = "Default binding behavior set to return Success"

		w.Write(res)

	default:
		res := ldapserver.NewResponse(ldapserver.LDAPResultUnwillingToPerform)
		res.DiagnosticMessage = "Operation not implemented by server"
		w.Write(res)
	}
}

func handleAbandon(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	var req = m.GetAbandonRequest()
	// retreive the request to abandon, and send a abort signal to it
	if requestToAbandon, ok := m.Client.GetMessageByID(int(req)); ok {
		requestToAbandon.Abandon()
		//logrus.Infof("Abandon signal sent to request processor [messageID=%d]", int(req))
	}
}

func handleBind(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	res := ldapserver.NewBindResponse(ldapserver.LDAPResultSuccess)
	r := m.GetBindRequest()
	//logrus.Debugf("Calling Bind Request for : User=%s, Pass=%#v", string(r.GetLogin()), string(r.GetPassword()))

	if string(r.GetPassword()) == TEST_ACCOUNT_VALID_PASSWORD {
		switch string(r.GetLogin()) {
		case fmt.Sprintf("uid=%s", GetTestAccountDN(TEST_ACCOUNT_TYPE_USER, TEST_ACCOUNT_ID_MAIN)),
			fmt.Sprintf("uid=%s", GetTestAccountDN(TEST_ACCOUNT_TYPE_USER, TEST_ACCOUNT_ID_1)),
			fmt.Sprintf("uid=%s", GetTestAccountDN(TEST_ACCOUNT_TYPE_USER, TEST_ACCOUNT_ID_2)),
			fmt.Sprintf("uid=%s", GetTestAccountDN(TEST_ACCOUNT_TYPE_USER, TEST_ACCOUNT_ID_3)),
			fmt.Sprintf("uid=%s,dc=example,dc=com", GetTestAccountDN(TEST_ACCOUNT_TYPE_USER, TEST_ACCOUNT_ID_MAIN)),
			fmt.Sprintf("uid=%s,dc=example,dc=com", GetTestAccountDN(TEST_ACCOUNT_TYPE_USER, TEST_ACCOUNT_ID_1)),
			fmt.Sprintf("uid=%s,dc=example,dc=com", GetTestAccountDN(TEST_ACCOUNT_TYPE_USER, TEST_ACCOUNT_ID_2)),
			fmt.Sprintf("uid=%s,dc=example,dc=com", GetTestAccountDN(TEST_ACCOUNT_TYPE_USER, TEST_ACCOUNT_ID_3)):
			w.Write(res)
			return
		}
	}

	//logrus.Debugf("Bind failed User=%s, Pass=%s", string(r.GetLogin()), string(r.GetPassword()))
	res.ResultCode = ldapserver.LDAPResultInvalidCredentials
	res.DiagnosticMessage = "invalid credentials"

	w.Write(res)
}

// The resultCode is set to compareTrue, compareFalse, or an appropriate
// error.  compareTrue indicates that the assertion value in the ava
// Comparerequest field matches a value of the attribute or subtype according to the
// attribute's EQUALITY matching rule.  compareFalse indicates that the
// assertion value in the ava field and the values of the attribute or
// subtype did not match.  Other result codes indicate either that the
// result of the comparison was Undefined, or that
// some error occurred.
func handleCompare(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	/*
		r := m.GetCompareRequest()
		logrus.Debugf("Comparing entry: %s", r.GetEntry())
		//attributes values
		logrus.Debugf(" attribute name to compare : \"%s\"", r.GetAttributeValueAssertion().GetName())
		logrus.Debugf(" attribute value expected : \"%s\"", r.GetAttributeValueAssertion().GetValue())
	*/

	res := ldapserver.NewCompareResponse(ldapserver.LDAPResultCompareTrue)
	w.Write(res)
}

func handleAdd(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	/*
		r := m.GetAddRequest()
		logrus.Debugf("Adding entry: %s", r.GetEntryDN())
		//attributes values
		for _, attribute := range r.GetAttributes() {
			for _, attributeValue := range attribute.GetValues() {
				logrus.Debugf("- %s:%s", attribute.GetDescription(), attributeValue)
			}
		}
	*/

	res := ldapserver.NewAddResponse(ldapserver.LDAPResultSuccess)
	w.Write(res)
}

func handleModify(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	/*
		r := m.GetModifyRequest()
		logrus.Debugf("Modify entry: %s", r.GetObject())

		for _, change := range r.GetChanges() {
			modification := change.GetModification()
			var operationString string
			switch change.GetOperation() {
			case ldapserver.ModifyRequestChangeOperationAdd:
				operationString = "Add"
			case ldapserver.ModifyRequestChangeOperationDelete:
				operationString = "Delete"
			case ldapserver.ModifyRequestChangeOperationReplace:
				operationString = "Replace"
			}


			logrus.Debugf("%s attribute '%s'", operationString, modification.GetDescription())
			for _, attributeValue := range modification.GetValues() {
				logrus.Debugf("- value: %s", attributeValue)
			}
		}
	*/

	res := ldapserver.NewModifyResponse(ldapserver.LDAPResultSuccess)
	w.Write(res)
}

func handleDelete(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	/*
		r := m.GetDeleteRequest()
		logrus.Debugf("Deleting entry: %s", r)
	*/
	res := ldapserver.NewDeleteResponse(ldapserver.LDAPResultSuccess)
	w.Write(res)
}

func handleExtended(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	/*
		r := m.GetExtendedRequest()
		logrus.Debugf("Extended request received, name=%s", r.GetResponseName())
		logrus.Debugf("Extended request received, value=%x", r.GetResponseValue())
	*/
	res := ldapserver.NewExtendedResponse(ldapserver.LDAPResultSuccess)
	w.Write(res)
}

func handleWhoAmI(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	res := ldapserver.NewExtendedResponse(ldapserver.LDAPResultSuccess)
	w.Write(res)
}

func handleSearchDSE(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	r := m.GetSearchRequest()

	attr := make([]string, 0)
	for _, v := range r.GetAttributes() {
		attr = append(attr, string(v))
	}

	/*
		logrus.Debugf("Request BaseDn=%s", r.GetBaseObject())
		logrus.Debugf("Request Filter=%s", r.GetFilter())
		logrus.Debugf("Request Attributes=%s", strings.Join(attr, ","))
		logrus.Debugf("Request TimeLimit=%d", r.GetTimeLimit())
	*/

	e := ldapserver.NewSearchResultEntry()
	e.AddAttribute("vendorName", "Test Vendor")
	e.AddAttribute("vendorVersion", "0.0.1")
	e.AddAttribute("objectClass", "top", "extensibleObject")
	e.AddAttribute("supportedLDAPVersion", "3")
	e.AddAttribute("namingContexts", "o=Test Company, c=US")
	w.Write(e)

	res := ldapserver.NewSearchResultDoneResponse(ldapserver.LDAPResultSuccess)
	w.Write(res)
}

func handleSearchMyCompany(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	r := m.GetSearchRequest()
	//logrus.Debugf("handleSearchMyCompany - Request BaseDn=%s", r.GetBaseObject())

	e := ldapserver.NewSearchResultEntry()
	e.SetDn(string(r.GetBaseObject()))
	e.AddAttribute("objectClass", "top", "organizationalUnit")
	w.Write(e)

	res := ldapserver.NewSearchResultDoneResponse(ldapserver.LDAPResultSuccess)
	w.Write(res)
}

func handleSearch(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	r := m.GetSearchRequest()

	if string(r.GetBaseObject()) == "" && r.GetScope() == ldapserver.SearchRequestScopeBaseObject && r.GetFilter() == "(objectclass=*)" {
		handleSearchDSE(w, m)
		return
	}

	if string(r.GetBaseObject()) == "o=My Company, c=US" && r.GetScope() == ldapserver.SearchRequestScopeBaseObject {
		handleSearchMyCompany(w, m)
		return
	}

	attr := make([]string, 0)
	for _, v := range r.GetAttributes() {
		attr = append(attr, string(v))
	}

	/*
		logrus.Debugf("Request BaseDn=%s", string(r.GetBaseObject()))
		logrus.Debugf("Request Filter=%s", r.GetFilter())
		logrus.Debugf("Request Attributes=%s", strings.Join(attr, ","))
		logrus.Debugf("Request TimeLimit=%d", r.GetTimeLimit())
	*/

	// Handle Stop Signal (server stop / client disconnected / Abandoned request....)
	select {
	case <-m.Done:
		//logrus.Info("Leaving handleSearch...")
		return
	default:
	}

	if r.GetFilter() == "(uid=Test1)" {
		//logrus.Debugf("Prepare Result Test1 for %s", r.GetFilter())
		e := ldapserver.NewSearchResultEntry()
		e.SetDn("uid=Test1," + string(r.GetBaseObject()))
		e.AddAttribute("mail", "test.ldap@example.com", "testldap@example.com")
		e.AddAttribute("uid", "Test1")
		e.AddAttribute("cn", "Test1")
		e.AddAttribute("ou", "People")
		e.AddAttribute("dc", "example", "com")
		e.AddAttribute("memberOf", "cn=Test - Group1,dc=example,dc=com", "cn=Test - Group2,dc=example,dc=com", "cn=Test - Group3,dc=example,dc=com")
		w.Write(e)
	}

	if r.GetFilter() == "(uid=Test2)" {
		//logrus.Debugf("Prepare Result Test2 for %s", r.GetFilter())
		e := ldapserver.NewSearchResultEntry()
		e.SetDn("uid=Test2," + string(r.GetBaseObject()))
		e.AddAttribute("mail", "test.2.ldap@example.com")
		e.AddAttribute("uid", "Test2")
		e.AddAttribute("cn", "Test2")
		e.AddAttribute("ou", "People")
		e.AddAttribute("dc", "example", "com")
		e.AddAttribute("memberOf", "cn=Test - Group1,dc=example,dc=com", "cn=Test - Group2,dc=example,dc=com")
		w.Write(e)
	}

	if r.GetFilter() == "(uid=Test3)" {
		//logrus.Debugf("Prepare Result Test3 for %s", r.GetFilter())
		e := ldapserver.NewSearchResultEntry()
		e.SetDn("uid=Test3," + string(r.GetBaseObject()))
		e.AddAttribute("mail", "test.3.ldap@example.com")
		e.AddAttribute("uid", "Test3")
		e.AddAttribute("cn", "Test3")
		e.AddAttribute("ou", "People")
		e.AddAttribute("dc", "example", "com")
		e.AddAttribute("memberOf", "cn=Test - Group1,dc=example,dc=com", "cn=Test - Group3,dc=example,dc=com")
		w.Write(e)
	}

	if r.GetFilter() == "(&(objectClass=groupOfNames)(cn=Test - Group1))" {
		//logrus.Debugf("Prepare Result [Test - Group1] for %s", r.GetFilter())
		e := ldapserver.NewSearchResultEntry()
		e.SetDn("uid=Test - Group1," + string(r.GetBaseObject()))
		e.AddAttribute("cn", "Test - Group1")
		e.AddAttribute("ou", "Group")
		e.AddAttribute("dc", "example", "com")
		e.AddAttribute("member", "uid=Test1,dc=example,dc=com", "uid=Test2,dc=example,dc=com", "uid=Test3,dc=example,dc=com")
		w.Write(e)
	}

	if r.GetFilter() == "(&(objectClass=groupOfNames)(cn=Test - Group2))" {
		//logrus.Debugf("Prepare Result [Test - Group2] for %s", r.GetFilter())
		e := ldapserver.NewSearchResultEntry()
		e.SetDn("uid=Test - Group2," + string(r.GetBaseObject()))
		e.AddAttribute("cn", "Test - Group2")
		e.AddAttribute("ou", "Group")
		e.AddAttribute("dc", "example", "com")
		e.AddAttribute("member", "uid=Test1,dc=example,dc=com", "uid=Test3,dc=example,dc=com")
		w.Write(e)
	}

	if r.GetFilter() == "(&(objectClass=groupOfNames)(cn=Test - Group3))" {
		//logrus.Debugf("Prepare Result [Test - Group3] for %s", r.GetFilter())
		e := ldapserver.NewSearchResultEntry()
		e.SetDn("uid=Test - Group3," + string(r.GetBaseObject()))
		e.AddAttribute("cn", "Test - Group3")
		e.AddAttribute("ou", "Group")
		e.AddAttribute("dc", "example", "com")
		e.AddAttribute("member", "uid=Test1,dc=example,dc=com", "uid=Test2,dc=example,dc=com")
		w.Write(e)
	}

	res := ldapserver.NewSearchResultDoneResponse(ldapserver.LDAPResultSuccess)

	//logrus.Debugf("Res Found : %v", res)
	w.Write(res)

}

// localhostCert is a PEM-encoded TLS cert with SAN DNS names
// "127.0.0.1" and "[::1]", expiring at the last second of 2049 (the end
// of ASN.1 time).
var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIIBOTCB5qADAgECAgEAMAsGCSqGSIb3DQEBBTAAMB4XDTcwMDEwMTAwMDAwMFoX
DTQ5MTIzMTIzNTk1OVowADBaMAsGCSqGSIb3DQEBAQNLADBIAkEAsuA5mAFMj6Q7
qoBzcvKzIq4kzuT5epSp2AkcQfyBHm7K13Ws7u+0b5Vb9gqTf5cAiIKcrtrXVqkL
8i1UQF6AzwIDAQABo08wTTAOBgNVHQ8BAf8EBAMCACQwDQYDVR0OBAYEBAECAwQw
DwYDVR0jBAgwBoAEAQIDBDAbBgNVHREEFDASggkxMjcuMC4wLjGCBVs6OjFdMAsG
CSqGSIb3DQEBBQNBAJH30zjLWRztrWpOCgJL8RQWLaKzhK79pVhAx6q/3NrF16C7
+l1BRZstTwIGdoGId8BRpErK1TXkniFb95ZMynM=
-----END CERTIFICATE-----
`)

// localhostKey is the private key for localhostCert.
var localhostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBPQIBAAJBALLgOZgBTI+kO6qAc3LysyKuJM7k+XqUqdgJHEH8gR5uytd1rO7v
tG+VW/YKk3+XAIiCnK7a11apC/ItVEBegM8CAwEAAQJBAI5sxq7naeR9ahyqRkJi
SIv2iMxLuPEHaezf5CYOPWjSjBPyVhyRevkhtqEjF/WkgL7C2nWpYHsUcBDBQVF0
3KECIQDtEGB2ulnkZAahl3WuJziXGLB+p8Wgx7wzSM6bHu1c6QIhAMEp++CaS+SJ
/TrU0zwY/fW4SvQeb49BPZUF3oqR8Xz3AiEA1rAJHBzBgdOQKdE3ksMUPcnvNJSN
poCcELmz2clVXtkCIQCLytuLV38XHToTipR4yMl6O+6arzAjZ56uq7m7ZRV0TwIh
AM65XAOw8Dsg9Kq78aYXiOEDc5DL0sbFUu/SlmRcCg93
-----END RSA PRIVATE KEY-----
`)

// getTLSconfig returns a tls configuration used
// to build a TLSlistener for TLS or StartTLS
func getTLSconfig() (*tls.Config, error) {
	cert, err := tls.X509KeyPair(localhostCert, localhostKey)
	if err != nil {
		return &tls.Config{}, err
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		ServerName:   "127.0.0.1",
	}, nil
}

func handleStartTLS(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	tlsconfig, e := getTLSconfig()
	if e != nil {
		logrus.Errorf("error while retrieve TLS config: %v", e)
	}

	tlsConn := tls.Server(m.Client.GetConn(), tlsconfig)
	res := ldapserver.NewExtendedResponse(ldapserver.LDAPResultSuccess)
	res.ResponseName = ldapserver.NoticeOfStartTLS
	w.Write(res)

	if err := tlsConn.Handshake(); err != nil {
		logrus.Errorf("StartTLS Handshake error %v", err)
		res.DiagnosticMessage = fmt.Sprintf("StartTLS Handshake error : \"%s\"", err.Error())
		res.ResultCode = ldapserver.LDAPResultOperationsError
		w.Write(res)
		return
	}

	m.Client.SetConn(tlsConn)
	logrus.Info("StartTLS OK")
}
