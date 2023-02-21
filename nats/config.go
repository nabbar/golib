/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package nats

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	moncfg "github.com/nabbar/golib/monitor/types"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	libiot "github.com/nabbar/golib/ioutils"
	liblog "github.com/nabbar/golib/logger"
	natjwt "github.com/nats-io/jwt/v2"
	natsrv "github.com/nats-io/nats-server/v2/server"
)

type Config struct {
	Server     ConfigSrv       `mapstructure:"server" json:"server" yaml:"server" toml:"server" validate:"dive,required"`
	Cluster    ConfigCluster   `mapstructure:"cluster" json:"cluster" yaml:"cluster" toml:"cluster" validate:"dive,required"`
	Gateways   ConfigGateway   `mapstructure:"gateways" json:"gateways" yaml:"gateways" toml:"gateways" validate:"dive,required"`
	Leaf       ConfigLeaf      `mapstructure:"leaf" json:"leaf" yaml:"leaf" toml:"leaf" validate:"dive,required"`
	Websockets ConfigWebsocket `mapstructure:"websockets" json:"websockets" yaml:"websockets" toml:"websockets" validate:"dive,required"`
	MQTT       ConfigMQTT      `mapstructure:"mqtt" json:"mqtt" yaml:"mqtt" toml:"mqtt" validate:"dive,required"`
	Limits     ConfigLimits    `mapstructure:"limits" json:"limits" yaml:"limits" toml:"limits" validate:"dive,required"`
	Logs       ConfigLogger    `mapstructure:"logs" json:"logs" yaml:"logs" toml:"logs" validate:"dive,required"`
	Auth       ConfigAuth      `mapstructure:"auth" json:"auth" yaml:"auth" toml:"auth" validate:"dive,required"`
	Monitor    moncfg.Config   `mapstructure:"monitor" json:"monitor" yaml:"monitor" toml:"monitor" validate:"dive"`

	//function / interface are not defined in config marshall
	Customs *ConfigCustom `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
}

func (c Config) Validate() liberr.Error {
	err := ErrorConfigValidation.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.AddParent(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

func (c Config) LogConfigJson() liberr.Error {
	if c.Logs.LogFile == "" {
		return nil
	}

	permFile := os.FileMode(0644)
	permDirs := os.FileMode(0755)

	if c.Logs.PermissionFileLogFile > 0 {
		permFile = c.Logs.PermissionFileLogFile
	}

	if c.Logs.PermissionFolderLogFile > 0 {
		permDirs = c.Logs.PermissionFolderLogFile
	}

	if e := libiot.PathCheckCreate(true, c.Logs.LogFile, permFile, permDirs); e != nil {
		return ErrorConfigInvalidFilePath.ErrorParent(e)
	}

	f, e := os.OpenFile(c.Logs.LogFile, os.O_APPEND|os.O_WRONLY, permFile)
	if e != nil {
		return ErrorConfigInvalidFilePath.ErrorParent(e)
	}

	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()

	if p, e := json.MarshalIndent(c, "", "  "); e != nil {
		return ErrorConfigJsonMarshall.ErrorParent(e)
	} else if _, e := f.WriteString("----\nConfig Node: "); e != nil {
		return ErrorConfigWriteInFile.ErrorParent(e)
	} else if _, e := f.Write(p); e != nil {
		return ErrorConfigWriteInFile.ErrorParent(e)
	} else if _, e := f.WriteString("\n---- \n"); e != nil {
		return ErrorConfigWriteInFile.ErrorParent(e)
	}

	return nil
}

func (c Config) NatsOption(defaultTls libtls.TLSConfig) (*natsrv.Options, liberr.Error) {
	cfg := &natsrv.Options{
		CheckConfig: false,
	}

	if e := c.Customs.makeOpt(cfg, defaultTls); e != nil {
		return nil, e
	}

	if e := c.Logs.makeOpt(cfg); e != nil {
		return nil, e
	}

	if e := c.Limits.makeOpt(cfg); e != nil {
		return nil, e
	}

	if e := c.Auth.makeOpt(cfg); e != nil {
		return nil, e
	}

	if e := c.Server.makeOpt(cfg, defaultTls); e != nil {
		return nil, e
	}

	if r, e := c.Cluster.makeOpt(defaultTls); e != nil {
		return nil, e
	} else {
		cfg.Cluster = r
	}

	if r, e := c.Gateways.makeOpt(defaultTls); e != nil {
		return nil, e
	} else {
		cfg.Gateway = r
	}

	if r, e := c.Leaf.makeOpt(cfg, c.Auth, defaultTls); e != nil {
		return nil, e
	} else {
		cfg.LeafNode = r
	}

	if r, e := c.Websockets.makeOpt(defaultTls); e != nil {
		return nil, e
	} else {
		cfg.Websocket = r
	}

	if r, e := c.MQTT.makeOpt(defaultTls); e != nil {
		return nil, e
	} else {
		cfg.MQTT = r
	}

	return cfg, nil
}

func (c *ConfigCustom) makeOpt(cfg *natsrv.Options, defTls libtls.TLSConfig) liberr.Error {
	if cfg == nil {
		return ErrorParamsInvalid.Error(nil)
	}

	if c == nil {
		return nil
	}

	if c.CustomClientAuthentication != nil {
		cfg.CustomClientAuthentication = c.CustomClientAuthentication
	}

	if c.CustomRouterAuthentication != nil {
		cfg.CustomRouterAuthentication = c.CustomRouterAuthentication
	}

	if c.AccountResolver != nil {
		cfg.AccountResolver = c.AccountResolver
	}

	if c.AccountResolverTLS {
		if t, e := c.AccountResolverTLSConfig.NewFrom(defTls); e != nil {
			return e
		} else {
			cfg.AccountResolverTLSConfig = t.TlsConfig("")
		}
	} else {
		cfg.AccountResolverTLSConfig = nil
	}

	return nil
}

func (c ConfigAuth) makeOpt(cfg *natsrv.Options) liberr.Error {
	if cfg == nil {
		return ErrorParamsInvalid.Error(nil)
	}

	if c.AuthTimeout > 0 {
		cfg.AuthTimeout = float64(c.AuthTimeout) / float64(time.Second)
	}

	if c.NoSystemAccount {
		cfg.NoSystemAccount = true
	}

	if c.SystemAccount != "" {
		cfg.SystemAccount = c.SystemAccount
	}

	if c.NoAuthUser != "" {
		cfg.NoAuthUser = c.NoAuthUser
	}

	if len(c.TrustedKeys) > 0 {
		cfg.TrustedKeys = c.TrustedKeys
	}

	if len(c.TrustedOperators) > 0 {
		cfg.TrustedOperators = make([]*natjwt.OperatorClaims, 0)

		for _, t := range c.TrustedOperators {
			if j, e := natsrv.ReadOperatorJWT(t); e != nil {
				return ErrorConfigInvalidJWTOperator.ErrorParent(e)
			} else if j != nil {
				cfg.TrustedOperators = append(cfg.TrustedOperators, j)
			}
		}
	}

	if len(c.NKeys) > 0 {
		cfg.Nkeys = make([]*natsrv.NkeyUser, 0)

		for _, k := range c.NKeys {
			if r, e := k.makeOpt(c, cfg); e != nil {
				return e
			} else if r != nil {
				cfg.Nkeys = append(cfg.Nkeys, r)
			}
		}
	}

	if len(c.Users) > 0 {
		cfg.Users = make([]*natsrv.User, 0)

		for _, k := range c.Users {
			if r, e := k.makeOpt(c, cfg); e != nil {
				return e
			} else if r != nil {
				cfg.Users = append(cfg.Users, r)
			}
		}
	}

	return nil
}

func (c ConfigNkey) makeOpt(auth ConfigAuth, cfg *natsrv.Options) (*natsrv.NkeyUser, liberr.Error) {
	if cfg == nil {
		return nil, ErrorParamsInvalid.Error(nil)
	}

	var (
		a *ConfigAccount
		t = make(map[string]struct{}, 0)
	)

	if c.Nkey == "" {
		return nil, nil
	}

	if c.SigningKey == "" {
		return nil, nil
	}

	if len(c.AllowedConnectionTypes) < 1 {
		c.AllowedConnectionTypes = []string{natjwt.ConnectionTypeStandard}
	}

	for _, at := range c.AllowedConnectionTypes {
		if at == "" {
			continue
		}
		switch strings.ToUpper(at) {
		case natjwt.ConnectionTypeStandard:
			t[natjwt.ConnectionTypeStandard] = struct{}{}
		case natjwt.ConnectionTypeWebsocket:
			t[natjwt.ConnectionTypeWebsocket] = struct{}{}
		case natjwt.ConnectionTypeLeafnode:
			t[natjwt.ConnectionTypeLeafnode] = struct{}{}
		case natjwt.ConnectionTypeMqtt:
			t[natjwt.ConnectionTypeMqtt] = struct{}{}
		default:
			return nil, ErrorConfigInvalidAllowedConnectionType.ErrorParent(fmt.Errorf("connection type: %s", at))
		}
	}

	if a = auth.findConfigAccount(c.Account); a == nil {
		return nil, ErrorConfigInvalidAccount.ErrorParent(fmt.Errorf("account: %s", c.Account))
	}

	return &natsrv.NkeyUser{
		Nkey: c.Nkey,
		Permissions: &natsrv.Permissions{
			Publish:   a.Permission.Publish.makeOpt(),
			Subscribe: a.Permission.Subscribe.makeOpt(),
			Response:  a.Permission.Response.makeOpt(),
		},
		Account:                auth.getAccount(cfg, c.Account),
		SigningKey:             c.SigningKey,
		AllowedConnectionTypes: t,
	}, nil
}

func (c ConfigUser) makeOpt(auth ConfigAuth, cfg *natsrv.Options) (*natsrv.User, liberr.Error) {
	if cfg == nil {
		return nil, ErrorParamsInvalid.Error(nil)
	}

	var (
		a *ConfigAccount
		t = make(map[string]struct{}, 0)
	)

	if c.Username == "" {
		return nil, nil
	}

	if c.Password == "" {
		return nil, nil
	}

	if len(c.AllowedConnectionTypes) < 1 {
		c.AllowedConnectionTypes = []string{natjwt.ConnectionTypeStandard}
	}

	for _, at := range c.AllowedConnectionTypes {
		if at == "" {
			continue
		}
		switch strings.ToUpper(at) {
		case natjwt.ConnectionTypeStandard:
			t[natjwt.ConnectionTypeStandard] = struct{}{}
		case natjwt.ConnectionTypeWebsocket:
			t[natjwt.ConnectionTypeWebsocket] = struct{}{}
		case natjwt.ConnectionTypeLeafnode:
			t[natjwt.ConnectionTypeLeafnode] = struct{}{}
		case natjwt.ConnectionTypeMqtt:
			t[natjwt.ConnectionTypeMqtt] = struct{}{}
		default:
			return nil, ErrorConfigInvalidAllowedConnectionType.ErrorParent(fmt.Errorf("connection type: %s", at))
		}
	}

	if a = auth.findConfigAccount(c.Account); a == nil {
		return nil, ErrorConfigInvalidAccount.ErrorParent(fmt.Errorf("account: %s", c.Account))
	}

	return &natsrv.User{
		Username: c.Username,
		Password: c.Password,
		Permissions: &natsrv.Permissions{
			Publish:   a.Permission.Publish.makeOpt(),
			Subscribe: a.Permission.Subscribe.makeOpt(),
			Response:  a.Permission.Response.makeOpt(),
		},
		Account:                auth.getAccount(cfg, c.Account),
		AllowedConnectionTypes: t,
	}, nil
}

func (c ConfigAuth) findConfigAccount(account string) *ConfigAccount {
	if len(c.Accounts) < 1 {
		return nil
	}

	for i, a := range c.Accounts {
		if a.Name == account {
			return &c.Accounts[i]
		}
	}

	return nil
}

func (c ConfigAuth) getAccount(cfg *natsrv.Options, account string) *natsrv.Account {
	a := natsrv.NewAccount(account)

	if len(cfg.Accounts) < 1 {
		cfg.Accounts = make([]*natsrv.Account, 0)
	}

	for i, n := range cfg.Accounts {
		if a.Name == n.Name {
			return cfg.Accounts[i]
		}
	}

	cfg.Accounts = append(cfg.Accounts, a)

	return a
}

func (c ConfigPermissionSubject) makeOpt() *natsrv.SubjectPermission {
	res := &natsrv.SubjectPermission{
		Allow: make([]string, 0),
		Deny:  make([]string, 0),
	}

	if len(c.Allow) > 0 {
		for _, p := range c.Allow {
			if p != "" {
				res.Allow = append(res.Allow, p)
			}
		}
	}

	if len(c.Deny) > 0 {
		for _, p := range c.Deny {
			if p != "" {
				res.Deny = append(res.Deny, p)
			}
		}
	}

	return res
}

func (c ConfigPermissionResponse) makeOpt() *natsrv.ResponsePermission {
	res := &natsrv.ResponsePermission{
		MaxMsgs: natsrv.DEFAULT_ALLOW_RESPONSE_MAX_MSGS,
		Expires: natsrv.DEFAULT_ALLOW_RESPONSE_EXPIRATION,
	}

	if c.MaxMsgs > 0 {
		res.MaxMsgs = c.MaxMsgs
	}

	if c.Expires > 0 {
		res.Expires = c.Expires
	}

	return res
}

func (c ConfigLogger) makeOpt(cfg *natsrv.Options) liberr.Error {
	if cfg == nil {
		return ErrorParamsInvalid.Error(nil)
	}

	var (
		permDir  os.FileMode = 0755
		permFile os.FileMode = 0644
	)

	if c.Syslog {
		cfg.Syslog = true
	}

	if c.RemoteSyslog != "" {
		cfg.RemoteSyslog = c.RemoteSyslog
	}

	if c.PermissionFolderLogFile > 0 {
		permDir = c.PermissionFolderLogFile
	}

	if c.PermissionFileLogFile > 0 {
		permFile = c.PermissionFileLogFile
	}

	if c.LogFile != "" {
		if e := libiot.PathCheckCreate(true, c.LogFile, permFile, permDir); e != nil {
			return ErrorConfigInvalidFilePath.ErrorParent(e)
		}
		cfg.LogFile = c.LogFile
	}

	if c.LogSizeLimit > 0 {
		cfg.LogSizeLimit = c.LogSizeLimit
	}

	if c.MaxTracedMsgLen > 0 {
		cfg.MaxTracedMsgLen = c.MaxTracedMsgLen
	}

	if c.ConnectErrorReports > 0 {
		cfg.ConnectErrorReports = c.ConnectErrorReports
	}

	if c.ReconnectErrorReports > 0 {
		cfg.ReconnectErrorReports = c.ReconnectErrorReports
	}

	if liblog.IsTimeStamp() {
		cfg.Logtime = true
	}

	if liblog.IsFileTrace() {
		cfg.Trace = true
	}

	switch liblog.GetCurrentLevel() {
	case liblog.DebugLevel:
		cfg.Debug = true
		cfg.NoLog = false
	case liblog.NilLevel:
		cfg.Debug = false
		cfg.NoLog = true
	default:
		cfg.Debug = false
		cfg.NoLog = false
	}

	return nil
}

func (c ConfigLimits) makeOpt(cfg *natsrv.Options) liberr.Error {
	if cfg == nil {
		return ErrorParamsInvalid.Error(nil)
	}

	if c.MaxConn > 0 {
		cfg.MaxConn = c.MaxConn
	}

	if c.MaxSubs > 0 {
		cfg.MaxSubs = c.MaxSubs
	}

	if c.PingInterval > 0 {
		cfg.PingInterval = c.PingInterval
	}

	if c.MaxPingsOut > 0 {
		cfg.MaxPingsOut = c.MaxPingsOut
	}

	if c.MaxControlLine > 0 {
		cfg.MaxControlLine = int32(c.MaxControlLine)
	}

	if c.MaxPayload > 0 {
		cfg.MaxPayload = int32(c.MaxPayload)
	}

	if c.MaxPending > 0 {
		cfg.MaxPending = c.MaxPending
	}

	if c.WriteDeadline > 0 {
		cfg.WriteDeadline = c.WriteDeadline
	}

	if c.MaxClosedClients > 0 {
		cfg.MaxClosedClients = c.MaxClosedClients
	}

	if c.LameDuckDuration > 0 {
		cfg.LameDuckDuration = c.LameDuckDuration
	}

	if c.LameDuckGracePeriod > 0 {
		cfg.LameDuckGracePeriod = c.LameDuckGracePeriod
	}

	if c.NoSublistCache {
		cfg.NoSublistCache = true
	}

	if c.NoHeaderSupport {
		cfg.NoHeaderSupport = true
	}

	if c.DisableShortFirstPing {
		cfg.DisableShortFirstPing = true
	}

	return nil
}

func (c ConfigSrv) makeOpt(cfg *natsrv.Options, defTls libtls.TLSConfig) liberr.Error {
	if cfg == nil {
		return ErrorParamsInvalid.Error(nil)
	}

	var (
		perm os.FileMode = 0755
	)

	if c.PermissionStoreDir > 0 {
		perm = c.PermissionStoreDir
	}

	if c.Name != "" {
		cfg.ServerName = c.Name
	}

	if c.Host != "" {
		cfg.Host = c.Host
	}

	if c.Port > 0 {
		cfg.Port = c.Port
	}

	if c.ClientAdvertise != "" {
		cfg.ClientAdvertise = c.ClientAdvertise
	}

	if c.HTTPHost != "" {
		cfg.HTTPHost = c.HTTPHost
	}

	if c.HTTPPort > 0 {
		cfg.HTTPPort = c.HTTPPort
	}

	if c.HTTPSPort > 0 {
		cfg.HTTPSPort = c.HTTPSPort
	}

	if c.HTTPBasePath != "" {
		cfg.HTTPBasePath = c.HTTPBasePath
	}

	if c.ProfPort > 0 {
		cfg.ProfPort = c.ProfPort
	}

	if c.PidFile != "" {
		cfg.PidFile = c.PidFile
	}

	if c.PortsFileDir != "" {
		cfg.PortsFileDir = c.PortsFileDir
	}

	if len(c.Routes) > 0 {
		cfg.Routes = make([]*url.URL, 0)

		for _, u := range c.Routes {
			if u == nil || u.Host == "" {
				continue
			}
			if u.Scheme == "" {
				u.Scheme = "nats"
			}
			cfg.Routes = append(cfg.Routes, u)
		}
	}

	if c.RoutesStr != "" {
		cfg.RoutesStr = c.RoutesStr
	}

	if c.NoSig {
		cfg.NoSigs = true
	}

	if c.Username != "" {
		cfg.Username = c.Username
	}

	if c.Password != "" {
		cfg.Password = c.Password
	}

	if c.Token != "" {
		cfg.Authorization = c.Token
	}

	if c.JetStream {
		cfg.JetStream = true

		if c.JetStreamMaxMemory > 0 {
			cfg.JetStreamMaxMemory = c.JetStreamMaxMemory
		}

		if c.JetStreamMaxStore > 0 {
			cfg.JetStreamMaxStore = c.JetStreamMaxStore
		}

		if c.StoreDir != "" {
			if e := libiot.PathCheckCreate(false, c.StoreDir, 0644, perm); e != nil {
				return ErrorConfigInvalidFilePath.ErrorParent(e)
			}

			cfg.StoreDir = c.StoreDir
		}
	}

	if len(c.Tags) > 0 {
		l := make(natjwt.TagList, 0)

		for _, t := range c.Tags {
			if t == "" {
				continue
			}
			l = append(l, t)
		}

		if len(l) > 0 {
			cfg.Tags = l
		}
	}

	if c.TLS {
		cfg.TLS = true

		if t, e := c.TLSConfig.NewFrom(defTls); e != nil {
			return e
		} else {
			cfg.TLSConfig = t.TlsConfig("")
		}

		if c.TLSTimeout > 0 {
			cfg.TLSTimeout = float64(c.TLSTimeout) / float64(time.Second)
		}

		if c.AllowNoTLS {
			cfg.AllowNonTLS = true
		}
	} else {
		cfg.TLS = false
		cfg.TLSConfig = nil
		cfg.TLSTimeout = 0
		cfg.HTTPSPort = 0
		cfg.AllowNonTLS = true
	}

	return nil
}

func (c ConfigCluster) makeOpt(defTls libtls.TLSConfig) (natsrv.ClusterOpts, liberr.Error) {
	cfg := natsrv.ClusterOpts{
		Name:              c.Name,
		Host:              c.Host,
		Port:              c.Port,
		Username:          c.Username,
		Password:          c.Password,
		AuthTimeout:       0,
		Permissions:       nil,
		TLSTimeout:        0,
		TLSConfig:         nil,
		TLSMap:            false,
		TLSCheckKnownURLs: false,
		ListenStr:         c.ListenStr,
		Advertise:         c.Advertise,
		NoAdvertise:       c.NoAdvertise,
		ConnectRetries:    c.ConnectRetries,
	}

	if c.AuthTimeout > 0 {
		cfg.AuthTimeout = float64(c.AuthTimeout) / float64(time.Second)
	}

	cfg.Permissions = &natsrv.RoutePermissions{
		Import: c.Permissions.Import.makeOpt(),
		Export: c.Permissions.Export.makeOpt(),
	}

	if c.TLS {
		if t, e := c.TLSConfig.NewFrom(defTls); e != nil {
			return cfg, e
		} else {
			cfg.TLSConfig = t.TlsConfig("")
		}

		if c.TLSTimeout > 0 {
			cfg.TLSTimeout = float64(c.TLSTimeout) / float64(time.Second)
		}
	} else {
		cfg.TLSConfig = nil
		cfg.TLSTimeout = 0
	}

	return cfg, nil
}

func (c ConfigGateway) makeOpt(defTls libtls.TLSConfig) (natsrv.GatewayOpts, liberr.Error) {
	cfg := natsrv.GatewayOpts{
		Name:              c.Name,
		Host:              c.Host,
		Port:              c.Port,
		Username:          c.Username,
		Password:          c.Password,
		AuthTimeout:       0,
		TLSConfig:         nil,
		TLSTimeout:        0,
		TLSMap:            false,
		TLSCheckKnownURLs: false,
		Advertise:         c.Advertise,
		ConnectRetries:    c.ConnectRetries,
		Gateways:          make([]*natsrv.RemoteGatewayOpts, 0),
		RejectUnknown:     c.RejectUnknown,
	}

	if c.AuthTimeout > 0 {
		cfg.AuthTimeout = float64(c.AuthTimeout) / float64(time.Second)
	}

	if c.TLS {
		if t, e := c.TLSConfig.NewFrom(defTls); e != nil {
			return cfg, e
		} else {
			cfg.TLSConfig = t.TlsConfig("")
		}

		if c.TLSTimeout > 0 {
			cfg.TLSTimeout = float64(c.TLSTimeout) / float64(time.Second)
		}
	}

	if len(c.Gateways) > 0 {
		for _, g := range c.Gateways {
			if r, e := g.makeOpt(defTls); e != nil {
				return cfg, e
			} else if r != nil {
				cfg.Gateways = append(cfg.Gateways, r)
			}
		}
	}

	return cfg, nil
}

func (c ConfigGatewayRemote) makeOpt(defTls libtls.TLSConfig) (*natsrv.RemoteGatewayOpts, liberr.Error) {
	res := &natsrv.RemoteGatewayOpts{
		Name:       "",
		TLSConfig:  nil,
		TLSTimeout: 0,
		URLs:       nil,
	}

	if c.Name != "" {
		res.Name = c.Name
	}

	if c.TLS {
		if t, e := c.TLSConfig.NewFrom(defTls); e != nil {
			return nil, e
		} else {
			res.TLSConfig = t.TlsConfig("")
		}

		if c.TLSTimeout > 0 {
			res.TLSTimeout = float64(c.TLSTimeout) / float64(time.Second)
		}
	} else {
		res.TLSConfig = nil
		res.TLSTimeout = 0
	}

	if len(c.URLs) > 0 {
		res.URLs = make([]*url.URL, 0)

		for _, u := range c.URLs {
			if u == nil || u.Host == "" {
				continue
			}
			res.URLs = append(res.URLs, u)
		}
	}

	return res, nil
}

func (c ConfigLeaf) makeOpt(cfg *natsrv.Options, auth ConfigAuth, defTls libtls.TLSConfig) (natsrv.LeafNodeOpts, liberr.Error) {
	res := natsrv.LeafNodeOpts{
		Host:              c.Host,
		Port:              c.Port,
		Username:          c.Username,
		Password:          c.Password,
		Account:           c.Account,
		Users:             make([]*natsrv.User, 0),
		AuthTimeout:       0,
		TLSConfig:         nil,
		TLSTimeout:        0,
		TLSMap:            false,
		Advertise:         c.Advertise,
		NoAdvertise:       c.NoAdvertise,
		ReconnectInterval: c.ReconnectInterval,
		Remotes:           make([]*natsrv.RemoteLeafOpts, 0),
	}

	if c.AuthTimeout > 0 {
		res.AuthTimeout = float64(c.AuthTimeout) / float64(time.Second)
	}

	if len(c.Users) > 0 {
		for _, u := range c.Users {
			if r, e := u.makeOpt(auth, cfg); e != nil {
				return res, e
			} else if r != nil {
				res.Users = append(res.Users, r)
			}
		}
	}

	if c.TLS {
		if t, e := c.TLSConfig.NewFrom(defTls); e != nil {
			return res, e
		} else {
			res.TLSConfig = t.TlsConfig("")
		}

		if c.TLSTimeout > 0 {
			res.TLSTimeout = float64(c.TLSTimeout) / float64(time.Second)
		}
	} else {
		res.TLSConfig = nil
		res.TLSTimeout = 0
	}

	if len(c.Remotes) > 0 {
		for _, l := range c.Remotes {
			if r, e := l.makeOpt(defTls); e != nil {
				return res, e
			} else if r != nil {
				res.Remotes = append(res.Remotes, r)
			}
		}
	}

	return res, nil
}

func (c ConfigLeafRemote) makeOpt(defTls libtls.TLSConfig) (*natsrv.RemoteLeafOpts, liberr.Error) {
	res := &natsrv.RemoteLeafOpts{
		LocalAccount: c.LocalAccount,
		URLs:         make([]*url.URL, 0),
		Credentials:  c.Credentials,
		TLS:          false,
		TLSConfig:    nil,
		TLSTimeout:   0,
		Hub:          c.Hub,
		DenyImports:  make([]string, 0),
		DenyExports:  make([]string, 0),
		Websocket: struct {
			Compression bool `json:"-"`
			NoMasking   bool `json:"-"`
		}{
			Compression: c.Websocket.Compression,
			NoMasking:   c.Websocket.NoMasking,
		},
	}

	if c.TLS {
		res.TLS = true

		if t, e := c.TLSConfig.NewFrom(defTls); e != nil {
			return nil, e
		} else {
			res.TLSConfig = t.TlsConfig("")
		}

		if c.TLSTimeout > 0 {
			res.TLSTimeout = float64(c.TLSTimeout) / float64(time.Second)
		}
	} else {
		res.TLS = false
		res.TLSConfig = nil
		res.TLSTimeout = 0
	}

	if len(c.URLs) > 0 {
		for _, u := range c.URLs {
			if u == nil || u.Host == "" {
				continue
			}
			res.URLs = append(res.URLs, u)
		}
	}

	if len(c.DenyImports) > 0 {
		res.DenyImports = c.DenyImports
	}

	if len(c.DenyExports) > 0 {
		res.DenyExports = c.DenyExports
	}

	return res, nil
}

func (c ConfigWebsocket) makeOpt(defTls libtls.TLSConfig) (natsrv.WebsocketOpts, liberr.Error) {
	cfg := natsrv.WebsocketOpts{
		Host:             c.Host,
		Port:             c.Port,
		Advertise:        c.Advertise,
		NoAuthUser:       c.NoAuthUser,
		JWTCookie:        c.JWTCookie,
		Username:         c.Username,
		Password:         c.Password,
		Token:            c.Token,
		AuthTimeout:      0,
		NoTLS:            false,
		TLSConfig:        nil,
		TLSMap:           false,
		SameOrigin:       c.SameOrigin,
		AllowedOrigins:   make([]string, 0),
		Compression:      c.Compression,
		HandshakeTimeout: 0,
	}

	if c.AuthTimeout > 0 {
		cfg.AuthTimeout = float64(c.AuthTimeout) / float64(time.Second)
	}

	if len(c.AllowedOrigins) > 0 {
		for _, o := range c.AllowedOrigins {
			if o != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, o)
			}
		}
	}

	if !c.NoTLS {
		cfg.NoTLS = false

		if t, e := c.TLSConfig.NewFrom(defTls); e != nil {
			return cfg, e
		} else {
			cfg.TLSConfig = t.TlsConfig("")
		}

		if c.HandshakeTimeout > 0 {
			cfg.HandshakeTimeout = c.HandshakeTimeout
		}
	} else {
		cfg.NoTLS = true
		cfg.TLSConfig = &tls.Config{}
		cfg.HandshakeTimeout = 0
	}

	return cfg, nil
}

func (c ConfigMQTT) makeOpt(defTls libtls.TLSConfig) (natsrv.MQTTOpts, liberr.Error) {
	cfg := natsrv.MQTTOpts{
		Host:          c.Host,
		Port:          c.Port,
		NoAuthUser:    c.NoAuthUser,
		Username:      c.Username,
		Password:      c.Password,
		Token:         c.Token,
		AuthTimeout:   0,
		TLSConfig:     nil,
		TLSMap:        false,
		TLSTimeout:    0,
		AckWait:       c.AckWait,
		MaxAckPending: c.MaxAckPending,
	}

	if c.AuthTimeout > 0 {
		cfg.AuthTimeout = float64(c.AuthTimeout) / float64(time.Second)
	}

	if !c.TLS {
		if t, e := c.TLSConfig.NewFrom(defTls); e != nil {
			return cfg, e
		} else {
			cfg.TLSConfig = t.TlsConfig("")
		}

		if c.TLSTimeout > 0 {
			cfg.TLSTimeout = float64(c.TLSTimeout) / float64(time.Second)
		}
	} else {
		cfg.TLSConfig = &tls.Config{}
		cfg.TLSTimeout = 0
	}

	return cfg, nil
}
